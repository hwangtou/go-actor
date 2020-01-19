package test

import (
	actor "github.com/go-actor"
	"log"
	"testing"
	"time"
)

func simple() {
	if err := actor.SetNewActorFn("simple_actor", newSimpleActor); err != nil {
		log.Println(err)
		return
	}
	aRef, err := actor.Spawn("simple_actor")
	if err != nil {
		log.Println(err)
		return
	}
	defer aRef.Release()
	if err := actor.Register(aRef.Id(), "simple_1"); err != nil {
		log.Println(err)
		return
	}
	log.Println(aRef)
	aRef.Tell(nil, "hello world", "extra")
}

func TestSimple(t *testing.T) {
	simple()
	time.After(time.Second)
}
