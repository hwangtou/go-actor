package go_actor

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrIdNotFound      = errors.New("id not found")
	ErrActorNotRunning = errors.New("actor halt")
	ErrActorName       = errors.New("actor name error")
	ErrActorCannotAsk  = errors.New("actor cannot ask")
)

const (
	actorBufferSize = 0
)

//
// Locals
//

type locals struct {
	actors      map[uint32]*LocalRef
	idCount     uint32
	idCountLock sync.Mutex
	names       map[string]uint32
	namesLock   sync.RWMutex
}

func (m *locals) init() {
	m.actors = map[uint32]*LocalRef{}
	m.idCount = 0
	m.idCountLock = sync.Mutex{}
	m.names = map[string]uint32{}
	m.namesLock = sync.RWMutex{}
}

// actors

func (m *locals) addActor(a Actor) *LocalRef {
	m.idCountLock.Lock()
	defer m.idCountLock.Unlock()

	for {
		m.idCount++
		if m.idCount == 0 {
			m.idCount++
		}
		if _, has := m.actors[m.idCount]; !has {
			break
		}
	}
	r := &LocalRef{}
	r.init(m.idCount, a, actorBufferSize)
	m.actors[m.idCount] = r
	return r
}

func (m *locals) getActor(id uint32) *LocalRef {
	r, has := m.actors[id]
	if !has {
		return nil
	}

	return r
}

func (m *locals) delActor(id uint32) {
	m.idCountLock.Lock()
	defer m.idCountLock.Unlock()

	delete(m.actors, id)
}

// actors life cycles

func (m *locals) spawnActor(actorType string, arg interface{}) (*LocalRef, error) {
	fn, has := sys.creators[actorType]
	if !has {
		return nil, ErrNewActorFnNotFound
	}
	a := fn()
	r := m.addActor(a)
	r.setStatusLock(ActorStartingUp)
	if err := a.StartUp(r, arg); err != nil {
		m.delActor(r.id.id)
		return nil, err
	}
	go r.spawn()
	return r, nil
}

func (m *locals) shutdownActor(ref *LocalRef) {
	// TODO to process global state
}

func (m *locals) haltActor(ref *LocalRef) {
	// TODO to process global state
	m.unbindName(ref)
	m.delActor(ref.id.id)
}

// names

func (m *locals) bindName(id uint32, name string) error {
	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	if name == "" {
		return ErrActorName
	}
	r, has := m.actors[id]
	if !has {
		return ErrIdNotFound
	}
	_, has = m.names[name]
	if has {
		return ErrActorNameExisted
	}
	r.id.name = name
	m.names[name] = id
	return nil
}

func (m *locals) unbindName(r *LocalRef) {
	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	if r.id.name != "" {
		delete(m.names, r.id.name)
		r.id.name = ""
	}
}

func (m *locals) getName(name string) *LocalRef {
	m.namesLock.RLock()
	defer m.namesLock.RUnlock()

	id, has := m.names[name]
	if !has {
		return nil
	}
	r, has := m.actors[id]
	if !has {
		return nil
	}

	return r
}

//
// LocalRef
//

type LocalRef struct {
	id         Id
	status     ActorStatus
	statusLock sync.RWMutex
	actor      Actor
	receiver   chan message
	executeAt  time.Time
	// sequence number counter
	sequence sequence
}

func (m *LocalRef) init(id uint32, a Actor, bufSize int) {
	m.id = Id{
		node: 0,
		id:   id,
		name: "",
	}
	m.status = ActorHalt
	m.statusLock = sync.RWMutex{}
	m.actor = a
	m.receiver = make(chan message, bufSize)
	m.executeAt = time.Time{}
	if ask, ok := a.(ActorAsk); ok {
		m.sequence.ask = ask
		m.sequence.sequences = map[uint64]sequenceWrapper{}
	}
}

func (m *LocalRef) setStatusLock(status ActorStatus) {
	m.statusLock.Lock()
	m.status = status
	m.statusLock.Unlock()
}

func (m *LocalRef) checkStatus(status ActorStatus) bool {
	m.statusLock.RLock()
	equal := m.status == status
	m.statusLock.RUnlock()
	return equal
}

func (m *LocalRef) logMessageError(err error, msg message) {
}

func (m *LocalRef) spawn() {
	m.setStatusLock(ActorRunning)
	for {
		msg := <-m.receiver
		m.executeAt = time.Now()
		log.Println(msg)
		switch msg.msgType {
		case msgTypeSend:
			m.actor.HandleSend(msg.sender, msg.msgContent)
		case msgTypeAsk:
			{
				// 1.Ask-Reply 2.Ask-Answer
				switch {
				case !msg.isReply && m.sequence.ask != nil:
					log.Println(m.id.name, ">>>>>>>>>>ask-reply", msg.sequence, m.sequence.sequences)
					answer, err := m.sequence.ask.HandleAsk(msg.sender, msg.msgContent)
					// TODO should I do this like that?
					go m.answer(message{
						sender:     msg.sender,
						ask:        m,
						sequence:   msg.sequence,
						msgType:    msgTypeAsk,
						msgContent: answer,
						msgError:   err,
						isReply:    true,
					})
				case msg.isReply:
					log.Println(m.id.name, ">>>>>>>>>>ask-answer", msg.sequence, m.sequence.sequences)
					w := m.sequence.getId(msg.sequence)
					if w == nil {
						// TODO
						break
					}
					w.messageCh <- msg
				default:
					// TODO
				}
			}
		case msgTypeKill:
			// TODO: There is a situation that cannot kill an actor:
			// the previous message is blocking this loop.
			{
				m.setStatusLock(ActorShuttingDown)
				sys.local.shutdownActor(m)
				if err := m.actor.Shutdown(); err != nil {
					m.logMessageError(err, msg)
				}
				m.actor = nil
				m.setStatusLock(ActorHalt)
				sys.local.haltActor(m)
				// Handle remaining
				// 1.Send 2.Ask 3.Ask-Reply 4.Kill
				for {
					msg, more := <-m.receiver
					if !more {
						break
					}
					m.logMessageError(ErrActorNotRunning, msg)
					switch {
					case msg.msgType == msgTypeSend:
					case msg.msgType == msgTypeAsk && !msg.isReply:
						// TODO To tell actor is not running
					case msg.msgType == msgTypeAsk && msg.isReply:
						// TODO To handle callback failed
					case msg.msgType == msgTypeKill:
					}
				}
				return
			}
		}
	}
}

func (m *LocalRef) Id() Id {
	return m.id
}

func (m *LocalRef) Send(sender Ref, msg interface{}) (err error) {
	// TODO critical state
	if !m.checkStatus(ActorRunning) {
		return ErrActorNotRunning
	}
	m.receiver <- message{
		sender:     sender,
		sequence:   0,
		msgType:    msgTypeSend,
		msgContent: msg,
		msgError:   nil,
		isReply:    false,
	}
	return nil
}

func (m *LocalRef) Ask(sender Ref, ask interface{}) (interface{}, error) {
	// TODO critical state
	if !m.checkStatus(ActorRunning) {
		return nil, ErrActorNotRunning
	}
	if m.sequence.ask == nil {
		return nil, ErrActorCannotAsk
	}
	wrapper := m.sequence.nextId()
	m.receiver <- message{
		sender:     sender,
		sequence:   wrapper.id,
		msgType:    msgTypeAsk,
		msgContent: ask,
		msgError:   nil,
		isReply:    false,
	}
	log.Println(m.id.name, ">>>>>>>>>>ask-ask", wrapper.id, m.sequence.sequences)
	resp := <-wrapper.messageCh
	return resp.msgContent, resp.msgError
}

func (m *LocalRef) Shutdown(sender Ref) error {
	// TODO critical state
	if !m.checkStatus(ActorRunning) {
		return ErrActorNotRunning
	}
	m.receiver <- message{
		sender:     sender,
		sequence:   0,
		msgType:    msgTypeKill,
		msgContent: nil,
		msgError:   nil,
		isReply:    false,
	}
	return nil
}

func (m *LocalRef) answer(reply message) {
	// TODO critical state
	if !m.checkStatus(ActorRunning) {
		// TODO
	}
	m.receiver <- reply
}
