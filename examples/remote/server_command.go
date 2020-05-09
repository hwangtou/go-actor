package main

import (
	"github.com/hwangtou/go-actor"
	"log"
	"os"
	"os/signal"
)

func main() {
	ch := make(chan os.Signal)
	log.Println("Actor system initializing")

	if err := actor.Remote.DefaultInit(1); err != nil {
		log.Fatalln("Actor system remote default init failed,", err)
	}

	if _, err := actor.SpawnWithName(newCommandActor, "command", nil); err != nil {
		log.Fatalln("Actor spawn name failed,", err)
	}

	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Stop actor system")
}

type command struct {
}

func newCommandActor() actor.Actor {
	return &command{}
}

func (m *command) Type() (name string, version int) {
	return "command", 1
}

func (m *command) StartUp(self *actor.LocalRef, arg interface{}) error {
	log.Println("Command actor is starting up")
	return nil
}

func (m *command) Started() {
	log.Println("Command actor has started")
}

func (m *command) HandleSend(sender actor.Ref, message interface{}) {
	log.Println("Command actor has receive send:", message)
}

func (m *command) Shutdown() {
	log.Println("Command actor is shutting down")
}

