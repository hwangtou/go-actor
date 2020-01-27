package test

import (
	actor "github.com/go-actor"
	"log"
	"testing"
	"time"
)

func simpleTwice1() *actor.LocalRef {
	aRef, err := actor.Spawn(constSimpleActor, nil)
	if err != nil {
		log.Fatalln(err)
	}
	if err := actor.Register(aRef.Id().ActorId(), "simple_1"); err != nil {
		log.Fatalln(err)
	}
	log.Println(aRef)
	return aRef
}

func simpleTwice2() *actor.LocalRef {
	aRef, err := actor.Spawn(constSimpleActor, nil)
	if err != nil {
		log.Fatalln(err)
	}
	if err := actor.Register(aRef.Id().ActorId(), "simple_2"); err != nil {
		log.Fatalln(err)
	}
	log.Println(aRef)
	return aRef
}

func simpleTwice() (*actor.LocalRef, *actor.LocalRef) {
	r1 := simpleTwice1()
	r2 := simpleTwice2()
	<-time.After(time.Second)
	if err := r1.Send(r2, "hello r1"); err != nil {
		log.Fatalln(err)
	}
	if err := r2.Send(r1, "hello r2"); err != nil {
		log.Fatalln(err)
	}
	return r1, r2
}

func TestSimpleTwice(t *testing.T) {
	r1, r2 := simpleTwice()
	if err := r2.Send(r1, "hello r2 again"); err != nil {
		log.Fatalln(err)
	}
	if answer, err := r2.Ask(r1, "testing ask"); err != nil {
		log.Fatalln(err)
	} else {
		log.Println(answer)
	}
	r1.Shutdown(nil)
	r2.Shutdown(nil)
	<-time.After(time.Second)
}
