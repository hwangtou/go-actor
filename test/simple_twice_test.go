package test

import (
	actor "github.com/go-actor"
	"log"
	"testing"
	"time"
)

func simpleTwice1() *actor.Ref {
	aRef, err := actor.Spawn(constSimpleActor)
	if err != nil {
		log.Println(err)
		return nil
	}
	if err := actor.Register(aRef.Id(), "simple_1"); err != nil {
		log.Println(err)
		return nil
	}
	log.Println(aRef)
	return aRef
}

func simpleTwice2() *actor.Ref {
	if err := actor.SetNewActorFn("simple_actor", newSimpleActor); err != nil {
		log.Println(err)
		return nil
	}
	aRef, err := actor.Spawn("simple_actor")
	if err != nil {
		log.Println(err)
		return nil
	}
	if err := actor.Register(aRef.Id(), "simple_2"); err != nil {
		log.Println(err)
		return nil
	}
	log.Println(aRef)
	return aRef
}

func simpleTwice() (*actor.Ref, *actor.Ref) {
	r1 := simpleTwice1()
	r2 := simpleTwice2()
	defer r1.Release()
	defer r2.Release()
	if err := r1.Send(r2.Id(), "hello r1"); err != nil {
		log.Println("eeeee1", err)
	}
	if err := r2.Send(r1.Id(), "hello r2"); err != nil {
		log.Println("eeeee2", err)
	}
	return r1, r2
}

func TestSimpleTwice(t *testing.T) {
	r1, r2 := simpleTwice()
	if err := r2.Send(r1.Id(), "hello r2 again"); err != nil {
		log.Println("eeeee3", err)
	}
	time.After(time.Second)
}
