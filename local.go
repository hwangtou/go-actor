package actor

import (
	"errors"
	"log"
	"reflect"
	"sync"
	"time"
)

const (
	// TODO 0 is for debug use, at least 1 in production
	actorBufferSize = 1
)

var (
	ErrAnswerType = errors.New("actor.Local answer type error")
	ErrMessageValue = errors.New("message value error")
)

//
// Locals
//

type localsManager struct {
	sys         *system
	sessions    sessionsManager
	actors      map[uint32]*LocalRef
	idCount     uint32
	idCountLock sync.Mutex
	names       map[string]nameWrapper
	namesLock   sync.RWMutex
}

func (m *localsManager) init(sys *system) {
	m.sys = sys
	m.sessions.init()
	m.actors = map[uint32]*LocalRef{}
	m.names = map[string]nameWrapper{}
}

// actor reference

func (m *localsManager) newActorRef(a Actor) *LocalRef {
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
	r.init(m, m.idCount, a, actorBufferSize)
	m.actors[m.idCount] = r
	return r
}

func (m *localsManager) getActorRef(id uint32) *LocalRef {
	r, has := m.actors[id]
	if !has {
		return nil
	}

	return r
}

func (m *localsManager) delActorRef(id uint32) {
	m.idCountLock.Lock()
	defer m.idCountLock.Unlock()

	delete(m.actors, id)
}

// actors life cycles

func (m *localsManager) spawnActor(fn func() Actor, name string, arg interface{}) (*LocalRef, error) {
	// #1 create actor with constructor function
	a := fn()
	// #2 new actor reference to hold created actor
	r := m.newActorRef(a)
	// #3 try starting up, lock the name if necessary
	if err := m.setNameSpawn(r, name, StartingUp); err != nil {
		return nil, err
	}
	r.setStatus(StartingUp)
	if err := a.StartUp(r, arg); err != nil {
		m.unsetNameSpawn(r, name, Halt)
		m.delActorRef(r.id.id)
		return nil, err
	}
	// #4 set running
	r.setStatus(Running)
	if err := m.setNameSpawn(r, name, Running); err != nil {
		m.unsetNameSpawn(r, name, Halt)
		m.delActorRef(r.id.id)
		return nil, err
	}
	// #5 SPAWN!!!
	go r.spawn()
	return r, nil
}

func (m *localsManager) shutdownActor(r *LocalRef) {
	name := r.id.name
	// #1 shutting down
	m.unsetNameSpawn(r, name, ShuttingDown)
	r.setStatus(ShuttingDown)
	close(r.recvCh)
	r.actor.Shutdown()
	r.actor = nil
	m.delActorRef(r.id.id)
	// #2 halt
	m.unsetNameSpawn(r, name, Halt)
	r.setStatus(Halt)
}

// names

type nameWrapper struct {
	id        uint32
	state     Status
	updatedAt time.Time
}

func (m *localsManager) setNameSpawn(ref *LocalRef, name string, status Status) error {
	if name == "" {
		return nil
	}

	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	n, has := m.names[name]
	switch status {
	case StartingUp:
		{
			// this branch will execute when actor is spawning with a name
			if has && n.state != Halt {
				return ErrNameRegistered
			}
			m.names[name] = nameWrapper{
				id:        ref.id.id,
				state:     status,
				updatedAt: time.Now(),
			}
			return nil
		}
	case Running:
		{
			ref.id.name = name
			m.names[name] = nameWrapper{
				id:        ref.id.id,
				state:     status,
				updatedAt: time.Now(),
			}
			return nil
		}
	}
	log.Panicf("set unsupported actor state:%v", status)
	return ErrActorState
}

func (m *localsManager) unsetNameSpawn(ref *LocalRef, name string, status Status) {
	if name == "" {
		return
	}
	if status != ShuttingDown && status != Halt {
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
		id:        0,
		state:     status,
		updatedAt: time.Now(),
	}
	return
}

func (m *localsManager) setNameRunning(ref *LocalRef, name string) error {
	if ref == nil || name == "" {
		return ErrArgument
	}
	if !ref.checkStatus(Running) {
		return ErrActorNotRunning
	}
	if ref.id.name != "" {
		return ErrNameRegistered
	}

	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	n, has := m.names[name]
	if has && n.state != Halt {
		return ErrActorState
	}
	ref.id.name = name
	m.names[name] = nameWrapper{
		id:        ref.id.id,
		state:     Running,
		updatedAt: time.Now(),
	}
	return nil
}

func (m *localsManager) getName(name string) *LocalRef {
	m.namesLock.RLock()
	defer m.namesLock.RUnlock()

	id, has := m.names[name]
	if !has {
		return nil
	}
	if id.state != Running {
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
	local       *localsManager
	id          Id
	status      Status
	statusLock  sync.RWMutex
	actor       Actor
	ask         Ask
	recvCh      chan *message
	recvRunning bool
	recvBeginAt time.Time
	recvEndAt   time.Time
}

func (m *LocalRef) init(local *localsManager, id uint32, a Actor, bufSize int) {
	m.local = local
	m.id = Id{
		node: 0,
		id:   id,
		name: "",
	}
	m.status = Halt
	m.actor = a
	if ask, ok := a.(Ask); ok {
		m.ask = ask
	}
	m.recvCh = make(chan *message, bufSize)
}

func (m *LocalRef) setStatus(status Status) {
	m.statusLock.Lock()
	m.status = status
	m.statusLock.Unlock()
}

func (m *LocalRef) checkStatus(status Status) bool {
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
			{
				m.actor.HandleSend(msg.sender, msg.msgContent)
			}
		case msgTypeAsk:
			{
				answer := message{
					sender:     msg.sender,
					msgSession: msg.msgSession,
					msgType:    msgTypeAnswer,
				}
				if m.ask == nil {
					answer.msgContent = nil
					answer.msgError = ErrActorCannotAsk
					m.local.sessions.handleSession(msg.msgSession, answer)
					break
				}
				answer.msgContent, answer.msgError = m.ask.HandleAsk(msg.sender, msg.msgContent)
				m.local.sessions.handleSession(msg.msgSession, answer)
			}
		case msgTypeKill:
			// TODO: There is a situation that cannot kill an actor:
			// the previous message is blocking this loop.
			{
				m.local.shutdownActor(m)
				m.recvRunning = false
				m.recvEndAt = time.Now()
				return
			}
		}
		m.recvRunning = false
		m.recvEndAt = time.Now()
	}
}

func (m LocalRef) Id() Id {
	return m.id
}

func (m *LocalRef) receiving(msg *message) (err error) {
	defer func() {
		if recover() != nil {
			log.Println("oops sending closed actor ref")
			err = ErrActorNotRunning
		}
	}()
	m.recvCh <- msg
	return nil
}

// #1 PLEASE DO NOT SEND VALUE CONTAINS chan, func, interface{}, pointer and unsafe
//    it will return ErrMessageValue
// #2 PLEASE DO NOT MODIFY SENT MESSAGE, no matter send side or receive side
//    it will affect the state of actor, especially MAP and ARRAY type!
func (m *LocalRef) Send(sender Ref, msg interface{}) (err error) {
	//if err := checkMessage(msg, false, 0); err != nil {
	//	return err
	//}
	// TODO critical state
	if !m.checkStatus(Running) {
		return ErrActorNotRunning
	}
	return m.receiving(&message{
		sender:     sender,
		msgSession: 0,
		msgType:    msgTypeSend,
		msgContent: msg,
		msgError:   nil,
	})
}

// #1 PLEASE DO NOT SEND VALUE CONTAINS chan, func, interface{}, pointer and unsafe
//    it will return ErrMessageValue
// #2 PLEASE DO NOT MODIFY SENT MESSAGE, no matter send side or receive side
//    it will affect the state of actor, especially MAP and ARRAY type!
func (m *LocalRef) Ask(sender Ref, ask interface{}, answer interface{}) error {
	answerValue := reflect.ValueOf(answer)
	if answerValue.Kind() != reflect.Ptr {
		return ErrAnswerType
	}
	// TODO critical state
	if !m.checkStatus(Running) {
		return ErrActorNotRunning
	}
	s := m.local.sessions.newSession()
	// sending to self
	if err := m.receiving(&message{
		sender:     sender,
		msgSession: s.id,
		msgType:    msgTypeAsk,
		msgContent: ask,
		msgError:   nil,
	}); err != nil {
		m.local.sessions.popSession(s.id)
		return err
	}
	// wait session to callback
	resp := <-s.msgCh
	if resp.msgContent == nil {
		e := answerValue.Elem()
		e.Set(reflect.Zero(e.Type()))
	} else {
		if !answerValue.Elem().Type().AssignableTo(reflect.ValueOf(resp.msgContent).Type()) {
			return ErrAnswerType
		}
		answerValue.Elem().Set(reflect.ValueOf(resp.msgContent))
	}
	return resp.msgError
}

func (m *LocalRef) Shutdown(sender Ref) error {
	// TODO critical state
	if !m.checkStatus(Running) {
		return ErrActorNotRunning
	}
	return m.receiving(&message{
		sender:     sender,
		msgSession: 0,
		msgType:    msgTypeKill,
		msgContent: nil,
		msgError:   nil,
	})
}

//
// message
//

type message struct {
	sender     Ref
	msgSession uint64
	msgType    messageType
	msgContent interface{}
	msgError   error
}

type messageType int

const (
	msgTypeSend   = 0
	msgTypeAsk    = 1
	msgTypeAnswer = 2
	msgTypeKill   = 3
)

// session manager

type sessionsManager struct {
	sync.Mutex
	currentId uint64
	sessions  map[uint64]*session
}

func (m *sessionsManager) init() {
	m.sessions = map[uint64]*session{}
}

func (m *sessionsManager) newSession() *session {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	for {
		m.currentId++
		if m.currentId == 0 {
			m.currentId++
		}
		if _, has := m.sessions[m.currentId]; !has {
			break
		}
	}
	w := &session{}
	w.init(m.currentId)
	m.sessions[m.currentId] = w
	return w
}

func (m *sessionsManager) popSession(id uint64) *session {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	w, has := m.sessions[id]
	if !has {
		return nil
	}
	delete(m.sessions, id)
	return w
}

func (m *sessionsManager) handleSession(id uint64, answer message) {
	s := m.popSession(id)
	if s == nil {
		log.Println("SERIOUS! Session not found!", id, answer)
		return
	}

	// TODO handle remote message
	defer func() {
		if recover() != nil {
			log.Println("oops handle closed session")
		}
	}()
	s.msgCh <- answer
}

// session wrapper

type session struct {
	id        uint64
	msgCh     chan message
	createdAt time.Time
}

func (m *session) init(id uint64) {
	m.id = id
	m.msgCh = make(chan message, 1)
	m.createdAt = time.Now()
}
