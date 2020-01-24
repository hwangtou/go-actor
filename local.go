package go_actor

import (
	"errors"
	"log"
	"sync"
)

var (
	ErrIdNotFound = errors.New("id not found")
	ErrActorHalt  = errors.New("actor halt")
	ErrActorName  = errors.New("actor name error")
	ErrActorCannotAsk = errors.New("actor cannot ask")
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
	m.idCount = 1
	m.idCountLock = sync.Mutex{}
	m.names = map[string]uint32{}
	m.namesLock = sync.RWMutex{}
}

// actors

func (m *locals) spawnActor(actorType string, arg interface{}) (*LocalRef, error) {
	fn, has := sys.creators[actorType]
	if !has {
		return nil, ErrNewActorFnNotFound
	}
	a := fn()
	r := m.addActor(a)
	if err := a.StartUp(r, arg); err != nil {
		m.delActor(r.id)
		return nil, err
	}
	return r, nil
}

func (m *locals) addActor(a Actor) *LocalRef {
	m.idCountLock.Lock()
	defer m.idCountLock.Unlock()

	for {
		if m.idCount == 0 {
			m.idCount += 1
		}
		if _, has := m.actors[m.idCount]; !has {
			break
		}
		m.idCount += 1
	}
	r := &LocalRef{
		node:      0,
		id:        m.idCount,
		name:      "",
		actor:     a,
		actorLock: sync.Mutex{},
		count:     1,
		countLock: sync.Mutex{},
		sequence: 1,
	}
	m.actors[m.idCount] = r
	return r
}

func (m *locals) getActor(id uint32) *LocalRef {
	r, has := m.actors[id]
	if !has {
		return nil
	}

	r.addCount()
	return r
}

func (m *locals) delActor(id uint32) {
	m.idCountLock.Lock()
	defer m.idCountLock.Unlock()

	delete(m.actors, id)
}

func (m *locals) shutdownActor(id uint32) error {
	r, has := m.actors[id]
	if !has {
		return ErrIdNotFound
	}
	m.unbindName(r)
	m.delActor(id)
	r.shutdown()
	return nil
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
	r.name = name
	m.names[name] = id
	return nil
}

func (m *locals) unbindName(r *LocalRef) {
	m.namesLock.Lock()
	defer m.namesLock.Unlock()

	if r.name != "" {
		delete(m.names, r.name)
		r.name = ""
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

	r.addCount()
	return r
}

//
// LocalRef
//

type LocalRef struct {
	node      uint32
	id        uint32
	name      string
	actor     Actor
	actorLock sync.Mutex
	count     int
	countLock sync.Mutex
	sequence  uint64
}

func (m *LocalRef) Id() *Id {
	return &Id{
		node: m.node,
		id:   m.id,
		name: m.name,
	}
}

func (m *LocalRef) Send(sender Ref, messages interface{}) (err error) {
	m.actorLock.Lock()
	defer m.actorLock.Unlock()

	if m.actor == nil {
		return ErrActorHalt
	}
	m.sending(&message{
		sender:     sender,
		sequence:   m.sequence,
		msgType:    msgTypeSend,
		msgContent: messages,
	})
	return nil
}

func (m *LocalRef) Ask(sender Ref, messages interface{}) (interface{}, error) {
	m.actorLock.Lock()
	defer m.actorLock.Unlock()

	if m.actor == nil {
		return nil, ErrActorHalt
	}
	reply := m.asking(&message{
		sender:     sender,
		sequence:   m.sequence,
		msgType:    msgTypeAsk,
		msgContent: messages,
	})
	return reply.msgContent, reply.msgError
}

func (m *LocalRef) sending(msg *message) {
	// TODO track running time
	m.actor.HandleSend(msg.sender, msg.msgContent)
}

func (m *LocalRef) asking(msg *message) *message {
	// TODO track running time
	a, ok := m.actor.(ActorAsk)
	if !ok {
		return &message{
			sender:     msg.sender,
			sequence:   msg.sequence,
			msgType:    msg.msgType,
			msgContent: msg.msgContent,
			msgError:   ErrActorCannotAsk,
		}
	}
	replies, err := a.HandleAsk(msg.sender, msg.msgContent)
	return &message{
		sender:     m,
		sequence:   msg.sequence,
		msgType:    msg.msgType,
		msgContent: replies,
		msgError:   err,
	}
}

func (m *LocalRef) addCount() {
	m.countLock.Lock()
	defer m.countLock.Unlock()

	m.count++
}

func (m *LocalRef) Retain() {
	m.addCount()
}

func (m *LocalRef) Release() {
	m.countLock.Lock()
	defer m.countLock.Unlock()

	if m.count > 0 {
		m.count -= 1
	}
	if m.count == 0 {
		m.actor.Idle()
	}
}

func (m *LocalRef) shutdown() {
	m.actorLock.Lock()
	defer m.actorLock.Unlock()

	if m.actor == nil {
		return
	}
	if err := m.actor.Shutdown(); err != nil {
		log.Println(err)
	}
	m.actor = nil
}
