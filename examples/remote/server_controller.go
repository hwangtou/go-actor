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

	// Config
	nodeConfig := actor.NodeConfig{
		Id:            2,
		ListenAddress: "127.0.0.1:12346",
	}
	remoteNodeDialConfig := actor.NodeConfig{
		Id:            1,
		ListenNetwork: actor.TCP4,
		ListenAddress: "127.0.0.1:12345",
		AuthToken:     "",
	}

	// Init node
	err := actor.Remote.Init(nodeConfig)
	if err != nil {
		log.Fatalln("Actor system remote default init failed,", err)
	}

	// Dialing to remote node
	conn, err := actor.Remote.Dial(remoteNodeDialConfig)
	if err != nil {
		log.Fatalln("Actor system get remote default failed,", err)
	}
	// Get reference of remote actor
	commandRef, err := conn.ByName("command")
	if err != nil {
		log.Println("Actor system get command actor failed,", err)
	} else {
		log.Println("Command ref,", commandRef)
		if err := commandRef.Send(nil, "it's ok"); err != nil {
			log.Println("Command send failed,", err)
		}
		answer := ""
		if err := commandRef.Ask(nil, "what's up", &answer); err != nil {
			log.Println("Command ask failed,", err)
		} else {
			log.Println("Command asked answer,", answer)
		}
	}

	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Stop actor system")
}
