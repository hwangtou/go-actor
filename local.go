package go_actor

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrActorNotRunning = errors.New("actor has halt")
	ErrActorState      = errors.New("actor state error")
	ErrArgument        = errors.New("argument error")
	ErrActorCannotAsk  = errors.New("actor cannot ask")
	ErrNameRegistered  = errors.New("name registered")
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
	names       map[string]nameWrapper
	namesLock   sync.RWMutex
	sessions    sessionManager
}

func (m *locals) init() {
	m.actors = map[uint32]*LocalRef{}
	m.idCount = 0
	m.names = map[string]nameWrapper{}
	m.sessions.init()
}

// actor reference

func (m *locals) newActorRef(a Actor) *LocalRef {
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

func (m *locals) getActorRef(id uint32) *LocalRef {
	r, has := m.actors[id]
	if !has {
		return nil
	}

	return r
}

func (m *locals) delActorRef(id uint32) {
	m.idCountLock.Lock()
	defer m.idCountLock.Unlock()

	delete(m.actors, id)
}

// actors life cycles

func (m *locals) spawnActor(fn ConstructorFn, name string, arg interface{}) (*LocalRef, error) {
	// #1 create actor with constructor function
	a := fn()
	// #2 new actor reference to hold created actor
	r := m.newActorRef(a)
	// #3 try starting up, lock the name if necessary
	if err := m.setNameSpawn(r, name, ActorStartingUp); err != nil {
		return nil, err
	}
	r.setStatus(ActorStartingUp)
	if err := a.StartUp(r, arg); err != nil {
		m.unsetNameSpawn(r, name, ActorHalt)
		m.delActorRef(r.id.id)
		return nil, ErrNameRegistered
	}
	// #4 set running
	r.setStatus(ActorRunning)
	if err := m.setNameSpawn(r, name, ActorRunning); err != nil {
		m.unsetNameSpawn(r, name, ActorHalt)
		m.delActorRef(r.id.id)
		return nil, ErrActorState
	}
	// #5 SPAWN!!!
	go r.spawn()
	return r, nil
}

func (m *locals) shutdownActor(r *LocalRef) {
	name := r.id.name
	// #1 shutting down
	m.unsetNameSpawn(r, name, ActorShuttingDown)
	r.setStatus(ActorShuttingDown)
	close(r.recvCh)
	r.actor.Shutdown()
	r.actor = nil
	m.delActorRef(r.id.id)
	// #2 halt
	m.unsetNameSpawn(r, name, ActorHalt)
	r.setStatus(ActorHalt)
}

// names

type nameWrapper struct {
	id        uint32
	state     ActorStatus
	updatedAt time.Time
}

func (m *locals) setNameSpawn(ref *LocalRef, name string, status ActorStatus) error {
	if name == "" {
		return nil
	}

	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	n, has := m.names[name]
	switch status {
	case ActorStartingUp:
		{
			// this branch will execute when actor is spawning with a name
			if has && n.state != ActorHalt {
				return ErrActorState
			}
			m.names[name] = nameWrapper{
				id:        ref.id.id,
				state:     status,
				updatedAt: time.Now(),
			}
			return nil
		}
	case ActorRunning:
		{
			ref.id.name = name
			m.names[name] = nameWrapper{
				id: ref.id.id,
				state: status,
				updatedAt: time.Now(),
			}
			return nil
		}
	}
	log.Panicf("set unsupported actor state:%v", status)
	return ErrActorState
}

func (m *locals) unsetNameSpawn(ref *LocalRef, name string, status ActorStatus) {
	if name == "" {
		return
	}
	if status != ActorShuttingDown && status != ActorHalt {
		log.Panicf("unset unsupported actor state:%v", status)
	}

	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	_, has := m.names[name]
	if !has {
		return
	}
	ref.id.name = ""
	m.names[name] = nameWrapper{
		id: 0,
		state: status,
		updatedAt: time.Now(),
	}
	return
}

func (m *locals) setNameRunning(ref *LocalRef, name string) error {
	if ref == nil || name == "" {
		return ErrArgument
	}
	if !ref.checkStatus(ActorRunning) {
		return ErrActorNotRunning
	}
	if ref.id.name != "" {
		return ErrNameRegistered
	}

	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	n, has := m.names[name]
	if has && n.state != ActorHalt {
		return ErrActorState
	}
	ref.id.name = name
	m.names[name] = nameWrapper{
		id:        ref.id.id,
		state:     ActorRunning,
		updatedAt: time.Now(),
	}
	return nil
}

func (m *locals) getName(name string) *LocalRef {
	m.namesLock.RLock()
	defer m.namesLock.RUnlock()

	id, has := m.names[name]
	if !has {
		return nil
	}
	if id.state != ActorRunning {
		return nil
	}
	r, has := m.actors[id.id]
	if !has {
		return nil
	}

	return r
}

//
// LocalRef
//

type LocalRef struct {
	id          Id
	status      ActorStatus
	statusLock  sync.RWMutex
	actor       Actor
	ask         ActorAsk
	recvCh      chan message
	recvRunning bool
	recvBeginAt time.Time
	recvEndAt   time.Time
}

func (m *LocalRef) init(id uint32, a Actor, bufSize int) {
	m.id = Id{
		node: 0,
		id:   id,
		name: "",
	}
	m.status = ActorHalt
	m.actor = a
	if ask, ok := a.(ActorAsk); ok {
		m.ask = ask
	}
	m.recvCh = make(chan message, bufSize)
}

func (m *LocalRef) setStatus(status ActorStatus) {
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

// TODO
func (m *LocalRef) logMessageError(err error, msg message) {
}

func (m *LocalRef) spawn() {
	m.actor.Started()
	for {
		// fetch new message
		msg := <-m.recvCh
		// mark recv time
		m.recvBeginAt = time.Now()
		m.recvRunning = true
		// handle
		switch msg.msgType {
		case msgTypeSend:
			m.actor.HandleSend(msg.sender, msg.msgContent)
		case msgTypeAsk:
			if m.ask == nil {
				// TODO handle error
				break
			}
			answer := message{
				sender:  msg.sender,
				askSeq:  msg.askSeq,
				msgType: msgTypeAnswer,
			}
			answer.msgContent, answer.msgError = m.ask.HandleAsk(msg.sender, msg.msgContent)
			// TODO handle global message
			sw := sys.local.sessions.getSession(msg.askSeq)
			if sw == nil {
				// TODO handle error
				break
			}
			sw.msgCh <- answer
		case msgTypeKill:
			// TODO: There is a situation that cannot kill an actor:
			// the previous message is blocking this loop.
			{
				sys.local.shutdownActor(m)
				m.recvRunning = false
				m.recvEndAt = time.Now()
				return
			}
		}
		m.recvRunning = false
		m.recvEndAt = time.Now()
	}
}

func (m *LocalRef) Id() Id {
	return m.id
}

func (m *LocalRef) receiving(msg message) (err error) {
	defer func() {
		if recover() != nil {
			log.Println("oops sending closed actor ref")
			err = ErrActorNotRunning
		}
	}()
	m.recvCh <- msg
	return nil
}

func (m *LocalRef) Send(sender Ref, msg interface{}) (err error) {
	// TODO critical state
	if !m.checkStatus(ActorRunning) {
		return ErrActorNotRunning
	}
	return m.receiving(message{
		sender:     sender,
		askSeq:     0,
		msgType:    msgTypeSend,
		msgContent: msg,
		msgError:   nil,
	})
}

func (m *LocalRef) Ask(sender Ref, ask interface{}) (interface{}, error) {
	// TODO critical state
	if !m.checkStatus(ActorRunning) {
		return nil, ErrActorNotRunning
	}
	sw := sys.local.sessions.genSession()
	if err := m.receiving(message{
		sender:     sender,
		askSeq:     sw.id,
		msgType:    msgTypeAsk,
		msgContent: ask,
		msgError:   nil,
	}); err != nil {
		sys.local.sessions.getSession(sw.id)
		return nil, err
	}
	resp := <-sw.msgCh
	return resp.msgContent, resp.msgError
}

func (m *LocalRef) Shutdown(sender Ref) error {
	// TODO critical state
	if !m.checkStatus(ActorRunning) {
		return ErrActorNotRunning
	}
	return m.receiving(message{
		sender:     sender,
		askSeq:     0,
		msgType:    msgTypeKill,
		msgContent: nil,
		msgError:   nil,
	})
}

func (m *LocalRef) IsLocalRunning() bool {
	return true
}

func (m *LocalRef) IsRemote() bool {
	if m.id.node == 0 {
		return false
	}
	return m.id.node != sys.node
}

func (m *LocalRef) IsGlobal() bool {
	return m.id.node != 0
}

//func (m *LocalRef) answer(reply message) {
//	// TODO critical state
//	if !m.checkStatus(ActorRunning) {
//		// TODO
//	}
//	m.recvCh <- reply
//}
