package test

import (
	actor "github.com/go-actor"
	"log"
	"testing"
	"time"
)

func simple() {
	aRef, err := actor.Spawn(constSimpleActor)
	if err != nil {
		log.Println(err)
		return
	}
	defer aRef.Release()
	if err := actor.Register(aRef.Id().ActorId(), "simple_1"); err != nil {
		log.Println(err)
		return
	}
	log.Println(aRef)
	aRef.Send(nil, "hello world 1")
	log.Println("hello world 1 next line")
	aRef.Send(nil, "hello world 2")
	log.Println("hello world 2 next line")
}

func TestSimple(t *testing.T) {
	simple()
	a := actor.ByName("simple_1")
	log.Println(a)
	time.After(time.Second)
}
