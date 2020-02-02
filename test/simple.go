package test

import (
	"fmt"
	actor "github.com/go-actor"
	"log"
	"time"
)

// SIMPLE ACTOR

type simpleActor struct {
	self actor.Ref
}

type signal int

func newSimpleActor() actor.Actor {
	return &simpleActor{}
}

func (m *simpleActor) Type() (name string, version int) {
	return "simple", 1
}

func (m *simpleActor) StartUp(self *actor.LocalRef, arg interface{}) error {
	log.Printf("start up id:%v\n", self.Id())
	m.self = self
	return nil
}

func (m *simpleActor) Started() {
	log.Printf("started id:%v\n", m.self.Id())
}

func (m *simpleActor) HandleSend(sender actor.Ref, message interface{}) {
	log.Printf("received raw:%v", message)
	senderId := "nil"
	if sender != nil {
		id := sender.Id()
		senderId = fmt.Sprintf("%d:%d:%s", id.NodeId(), id.ActorId(), id.Name())
	}
	switch msg := message.(type) {
	case *simpleActorMessage:
		log.Printf("received message:%v from %v\n", msg, senderId)
	case signal:
		log.Printf("received signal:%v from %v\n", msg, senderId)
	case string:
		log.Printf("received string:%v from %v\n", msg, senderId)
	case error:
		log.Printf("received error:%v from %v\n", msg, senderId)
	default:
		log.Printf("received unknown:%v\n", msg)
	}
	<-time.After(time.Second * 1)
}

func (m *simpleActor) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	log.Printf("received ask:%v\n", ask)
	return "ANSWER", nil
}

func (m *simpleActor) Shutdown() {
	log.Println("shutdown")
}

type simpleActorMessage struct {
}
