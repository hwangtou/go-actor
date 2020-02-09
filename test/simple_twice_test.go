package test

import (
	actor "github.com/go-actor"
	"log"
	"testing"
	"time"
)

type a struct {
}

func simpleTwice1() *actor.LocalRef {
	aRef, err := actor.Spawn(func() actor.Actor {
		return &simpleActor{}
	}, nil)
	if err != nil {
		log.Fatalln(err)
	}
	if err := actor.Register(aRef, "simple_1"); err != nil {
		log.Fatalln(err)
	}
	log.Println(aRef)
	return aRef
}

func simpleTwice2() *actor.LocalRef {
	aRef, err := actor.SpawnWithName(newSimpleActor, "simple_2", nil)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(aRef)
	return aRef
}

func simpleTwice() (*actor.LocalRef, *actor.LocalRef) {
	r1 := simpleTwice1()
	r2 := simpleTwice2()
	<-time.After(time.Second)
	if err := r1.Send(r2, struct{}{}); err != nil {
		log.Println("message error: struct{}{}", err)
	}
	if err := r1.Send(r2, nil); err != nil {
		log.Println("message error: nil", err)
	}
	if err := r1.Send(r2, a{}); err != nil {
		log.Println("message error: a{}", err)
	}
	log.Println("sending &a{}")
	var v *a = &a{}
	if err := r1.Send(r2, v); err != nil {
		log.Println("message error: &a{}", err)
	}
	if err := r2.Send(r1, "hello r2"); err != nil {
		log.Println("message error: string", err)
	}
	return r1, r2
}

func TestSimpleTwice(t *testing.T) {
	r1, r2 := simpleTwice()
	if err := r2.Send(r1, "hello r2 again"); err != nil {
		log.Println("send error: string", err)
	}
	var a1 simpleActorMessage
	if err := r2.Ask(r1, "struct", &a1); err != nil {
		log.Println("ask error: struct", err)
	} else {
		log.Println("answer 1", a1)
	}

	var a2 int
	if err := r2.Ask(r1, "int", &a2); err != nil {
		log.Println("ask error: int", err)
	} else {
		log.Println("answer 2", a2)
	}

	var a3 *simpleActorMessage
	if err := r2.Ask(r1, "structPtr", &a3); err != nil {
		log.Println("ask error: struct", err)
	} else {
		log.Println("answer 3", a3)
	}

	var a4 *simpleActorMessage
	if err := r2.Ask(r1, "nil", &a4); err != nil {
		log.Println("ask error: nil", a4, ", err", err)
	} else {
		log.Println("answer 4", a4)
	}

	r1.Shutdown(nil)
	r2.Shutdown(nil)
	<-time.After(1 * time.Second)
}
