package test

import (
	"github.com/hwangtou/go-actor"
	"log"
	"testing"
	"time"
)

func simple() {
	aRef, err := actor.Spawn(func() actor.Actor { return &simpleActor{} }, nil)
	if err != nil {
		log.Fatalln(err)
	}
	if err := actor.Register(aRef, "simple_1"); err != nil {
		log.Fatalln(err)
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
	a.Shutdown(nil)
	<-time.After(time.Second * 2)
}
