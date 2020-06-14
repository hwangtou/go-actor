package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hwangtou/go-actor"
	"github.com/hwangtou/go-actor/handlers/websocket"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const dialerName = "Dialer"
const serverAddr = "ws://127.0.0.1:12345/websocket"

func newDialer() actor.Actor {
	return &websocket.Dialer{}
}

func conn(dialerRef *actor.LocalRef, name, password string) {
	// Check existed
	connForwarder := actor.ByName(name)
	if connForwarder != nil {
		fmt.Println("Connected")
		return
	}
	// Dial to server
	var connRef *actor.LocalRef
	header := http.Header{}
	header.Set("password", password)
	ask := &websocket.Dialing{
		Url:           serverAddr,
		ForwardRef:    dialerRef,
		RequestHeader: header,
		ReadTimeout:   time.Time{},
		WriteTimeout:  time.Time{},
	}
	if err := dialerRef.Ask(nil, ask, &connRef); err != nil {
		fmt.Println("Failed to dial server, ", err)
		return
	}
	// Spawn conn forwarder actor
	if _, err := actor.SpawnWithName(newConnForwarder, name, &NewConnArgs{
		ConnRef: connRef,
	}); err != nil {
		fmt.Println("Failed to spawn conn forwarder, ", err)
	}
}

func send(name, message string) {
	// Check existed
	connForwarder := actor.ByName(name)
	if connForwarder == nil {
		fmt.Println("Not connected")
		return
	}
	if err := connForwarder.Send(nil, message); err != nil {
		fmt.Println("Send error, ", err)
	}
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
			send(name, message)
		default:
			fmt.Println("Invalid Command: " + commands[0])
		}
	}
exit:
	log.Println("Exiting dialer")
}

// Conn Forwarder Actor

type connForwarder struct {
	self    *actor.LocalRef
	connRef actor.Ref
}

type NewConnArgs struct {
	ConnRef actor.Ref
}

func newConnForwarder() actor.Actor {
	return &connForwarder{}
}

func (m *connForwarder) Type() (name string, version int) {
	return "ConnForwarder", 1
}

func (m *connForwarder) StartUp(self *actor.LocalRef, arg interface{}) error {
	newConnArg, ok := arg.(*NewConnArgs)
	if !ok {
		return errors.New("invalid StartUp arg")
	}
	m.self = self
	m.connRef = newConnArg.ConnRef
	return nil
}

func (m *connForwarder) Started() {
}

func (m *connForwarder) HandleSend(sender actor.Ref, message interface{}) {
	switch message.(type) {
	case *websocket.ReceiveClosed:
		fmt.Println("Received closed")
	}
}

func (m *connForwarder) Shutdown() {
}
