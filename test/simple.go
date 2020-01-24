package test

import (
	actor "github.com/go-actor"
	"log"
	"time"
)

// SIMPLE ACTOR

const constSimpleActor = "simple_actor"

func init() {
	if err := actor.SetNewActorFn(constSimpleActor, newSimpleActor); err != nil {
		log.Println(err)
		return
	}
}

type simpleActor struct {
	self actor.Ref
}

type signal int

func newSimpleActor() actor.Actor {
	return &simpleActor{}
}

func (m *simpleActor) StartUp(self actor.Ref, arg interface{}) error {
	log.Printf("start up id:%v\n", self.Id())
	m.self = self
	return nil
}

func (m *simpleActor) HandleSend(sender actor.Ref, message interface{}) {
	log.Printf("received raw:%v", message)
	switch msg := message.(type) {
	case *simpleActorMessage:
		log.Printf("received message:%v from %v\n", msg, sender)
	case signal:
		log.Printf("received signal:%v from %v\n", msg, sender)
	case string:
		log.Printf("received string:%v from %v\n", msg, sender)
	case error:
		log.Printf("received error:%v from %v\n", msg, sender)
	default:
		log.Printf("received unknown:%v\n", msg)
	}
	<-time.After(time.Second * 1)
}

func (m *simpleActor) Idle() {
	log.Println("idle, to shutdown")
	if err := actor.ShutDown(m.self.Id().ActorId()); err != nil {
		log.Println(err)
	}
	if err := actor.ShutDown(m.self.Id().ActorId()); err != nil {
		log.Println(err)
	}
}

func (m *simpleActor) Shutdown() error {
	log.Println("shutdown")
	return nil
}

type simpleActorMessage struct {
}
