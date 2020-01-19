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
	actorIdIter uint32
	names       map[string]uint32
}

func (m *manager) init() {
	m.actors = map[uint32]*Ref{}
	m.actorIdIter = 1
	m.names = map[string]uint32{}
}

func (m *manager) nextId() uint32 {
	for {
		if m.actorIdIter == 0 {
			m.actorIdIter += 1
		}
		if _, has := m.actors[m.actorIdIter]; !has {
			break
		}
		m.actorIdIter += 1
	}
	return m.actorIdIter
}

func (m *manager) newActor(a Actor) *Ref {
	id := m.nextId()
	r := &Ref{
		node:      0,
		id:        id,
		name:      "",
		actor:     a,
		actorLock: sync.Mutex{},
		count:     1,
		countLock: sync.Mutex{},
	}
	m.actors[id] = r
	return r
}

func (m *manager) bind(id uint32, name string) error {
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

func (m *manager) unbind(r *Ref) {
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
}

func (m *Ref) Id() *Id {
	return &Id{
		node: m.node,
		id:   m.id,
		name: "",
	}
}

func (m *Ref) Tell(sender *Id, message ...interface{}) (err error) {
	m.actorLock.Lock()
	defer m.actorLock.Unlock()

	if m.actor == nil {
		return ErrActorHalt
	}
	go m.actor.HandleTell(sender, message...)
	return nil
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
