package main

import (
	actor "github.com/go-actor"
	"github.com/go-actor/demo/chat"
	"log"
)

func main() {
	if err := actor.SetNewActorFn(chat.NewRoomManagerKey, chat.NewRoomManager); err != nil {
		log.Println("set new actor fn, ", err)
		return
	}
	if err := actor.SetNewActorFn(chat.NewRoomKey, chat.NewRoom); err != nil {
		log.Println("set new actor fn, ", err)
		return
	}
	if err := actor.SetNewActorFn(chat.NewClientManagerKey, chat.NewClientManager); err != nil {
		log.Println("set new actor fn, ", err)
		return
	}
	if err := actor.SetNewActorFn(chat.NewClientKey, chat.NewClient); err != nil {
		log.Println("set new actor fn, ", err)
		return
	}
	_, err := actor.Spawn(chat.NewRoomManagerKey, nil)
	if err != nil {
		log.Println("spawn failed, ", err)
		return
	}
	clientManagerRef, err := actor.Spawn(chat.NewClientManagerKey, "127.0.0.1:20000")
	if err != nil {
		log.Println("spawn failed, ", err)
		return
	}

	exit := make(chan bool)
	<-exit

	clientManagerRef.Release()
}
