package main

import (
	"bufio"
	"fmt"
	"github.com/hwangtou/go-actor"
	"github.com/hwangtou/go-actor/handlers/websocket"
	"log"
	"os"
	"strings"
)

const dialerName = "Dialer"

func newDialer() actor.Actor {
	return &websocket.Dialer{}
}

func conn(dialer *actor.LocalRef, name, password string) {

}

func send(dialer *actor.LocalRef, name, message string) {

}

func main() {
	dialerRef, err := actor.SpawnWithName(newDialer, dialerName, nil)
	if err != nil {
		log.Fatalln("Spawning dialer error,", err)
	}

	log.Println("Wait for command, conn>name:password, send>name:message")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Command: ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		text = strings.Replace(text, "\r", "", -1)
		commands := strings.Split(text, ">")
		if len(commands) < 1 {
			fmt.Println("Invalid Command: " + text)
			continue
		}
		switch commands[0] {
		case "exit":
			fmt.Println("Exiting")
			goto exit
		case "conn":
			fmt.Println("Connecting to conn")
			args := strings.Split(commands[1], ":")
			if len(args) < 2 {
				fmt.Println("Invalid Command: " + commands[1])
				continue
			}
			name, password := args[0], args[1]
			conn(dialerRef, name, password)
		case "send":
			fmt.Println("Send to conn")
			args := strings.Split(commands[1], ":")
			if len(args) < 2 {
				fmt.Println("Invalid Command: " + commands[1])
				continue
			}
			name, message := args[0], args[1]
			send(dialerRef, name, message)
		default:
			fmt.Println("Invalid Command: " + commands[0])
		}
	}
exit:
	log.Println("Exiting dialer")
}
