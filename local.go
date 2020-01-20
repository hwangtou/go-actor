package go_actor

import (
	"errors"
	"log"
	"sync"
)

var (
	ErrIdNotFound = errors.New("id not found")
	ErrActorHalt  = errors.New("id halt")
	ErrActorName  = errors.New("id name error")
)

//
// Manager
//

type manager struct {
	actors      map[uint32]*Ref
	idCount     uint32
	idCountLock sync.Mutex
	names       map[string]uint32
}

func (m *manager) init() {
	m.actors = map[uint32]*Ref{}
	m.idCount = 1
	m.idCountLock = sync.Mutex{}
	m.names = map[string]uint32{}
}

func (m *manager) addActor(a Actor) *Ref {
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
	r := &Ref{
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

func (m *manager) delActor(id uint32) {
	m.idCountLock.Lock()
	defer m.idCountLock.Unlock()

	delete(m.actors, id)
}

func (m *manager) bindName(id uint32, name string) error {
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

func (m *manager) unbindName(r *Ref) {
	if r.name != "" {
		delete(m.names, r.name)
		r.name = ""
	}
}

//
// Ref
//

type Ref struct {
	node      uint32
	id        uint32
	name      string
	actor     Actor
	actorLock sync.Mutex
	count     int
	countLock sync.Mutex
	sequence  uint64
}

func (m *Ref) Id() *Id {
	return &Id{
		node: m.node,
		id:   m.id,
		name: "",
	}
}

func (m *Ref) Send(sender *Id, messages ...interface{}) (err error) {
	m.actorLock.Lock()
	if m.actor == nil {
		defer m.actorLock.Unlock()
		return ErrActorHalt
	}
	go func(sender *Id, messages ...interface{}) {
		defer m.actorLock.Unlock()
		m.handleMessage(&message{
			sender:   sender,
			sequence: m.sequence,
			messages: messages,
		})
	}(sender, messages...)
	return nil
}

func (m *Ref) handleMessage(msg *message) {
	// TODO track running time
	m.actor.HandleSend(msg.sender, msg.messages...)
}

func (m *Ref) Release() {
	m.countLock.Lock()
	defer m.countLock.Unlock()

	if m.count > 0 {
		m.count -= 1
	}
	if m.count == 0 {
		m.actor.Idle()
	}
}

func (m *Ref) shutdown() {
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

//
// message
//

type message struct {
	sender *Id
	sequence uint64
	messages []interface{}
}

//
// Id
//

type Id struct {
	node, id uint32
	name     string
}

func (m *Id) NodeId() uint32 {
	return m.node
}

func (m *Id) ActorId() uint32 {
	return m.id
}
