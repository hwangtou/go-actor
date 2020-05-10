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

	if err := actor.Remote.Init(actor.NodeConfig{
		Id:            2,
		ListenAddress: "127.0.0.1:12346",
	}); err != nil {
		log.Fatalln("Actor system remote default init failed,", err)
	}

	conn, err := actor.Remote.NewConn(1, "", "", actor.TCP4, "127.0.0.1:12345")
	if err != nil {
		log.Fatalln("Actor system get remote default failed,", err)
	}
	commandRef, err := conn.ByName("command")
	if err != nil {
		log.Println("Actor system get command actor failed,", err)
	} else {
		log.Println("Command ref,", commandRef)
		if err := commandRef.Send(nil, "it's ok"); err != nil {
			log.Println("Command send failed,", err)
		}
	}

	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Stop actor system")
}
