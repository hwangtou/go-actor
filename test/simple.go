package test

import (
	actor "github.com/go-actor"
	"log"
)

// SIMPLE ACTOR

type simpleActor struct {
	id uint32
}

type signal int

func newSimpleActor() actor.Actor {
	return &simpleActor{}
}

func (m *simpleActor) StartUp(id uint32) error {
	log.Printf("start up id:%v\n", id)
	m.id = id
	return nil
}

func (m *simpleActor) HandleTell(sender *actor.Id, messages ...interface{}) {
	log.Printf("received raw:%v", messages)
	switch msg := messages[0].(type) {
	case signal:
		log.Printf("received signal:%v from %v\n", msg, sender)
	case string:
		log.Printf("received string:%v from %v\n", msg, sender)
	case error:
		log.Printf("received error:%v from %v\n", msg, sender)
	default:
		log.Printf("received unknown:%v\n", msg)
	}
}

func (m *simpleActor) Idle() {
	log.Println("idle, to shutdown")
	if err := actor.ShutDown(m.id); err != nil {
		log.Println(err)
	}
	if err := actor.ShutDown(m.id); err != nil {
		log.Println(err)
	}
}

func (m *simpleActor) Shutdown() error {
	log.Println("shutdown")
	return nil
}
