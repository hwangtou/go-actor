package main

import (
	"errors"
	"fmt"
	"github.com/hwangtou/go-actor"
	"github.com/hwangtou/go-actor/handlers/websocket"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	ch := make(chan os.Signal)
	log.Println("Spawning auth forwarder actor")
	if _, err := actor.SpawnWithName(newAuthForwarder, authForwarderName, nil); err != nil {
		log.Println(err)
	}
	log.Println("Spawning listener actor")
	if _, err := actor.SpawnWithName(newListener, listenerName, &websocket.StartUpConfig{
		IsProduction:  true,
		ListenNetwork: websocket.TCP4,
		ListenAddr:    listenerAddr,
		TLS:           nil,
		Handlers: []websocket.HandlerConfig{
			{
				Method:           websocket.GET,
				RelativePath:     "/websocket",
				ForwardActorName: authForwarderName,
				ReadTimeout:      time.Time{},
				WriteTimeout:     time.Time{},
			},
		},
	}); err != nil {
		log.Fatalln(err)
	}

	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Stop listener")
}

//
// Auth Forwarder Actor
//

type authForwarder struct {
	self *actor.LocalRef
}

const authForwarderName = "AuthForwarder"

func newAuthForwarder() actor.Actor {
	return &authForwarder{}
}

func (m *authForwarder) Type() (name string, version int) {
	return "AuthForwarder", 1
}

func (m *authForwarder) StartUp(self *actor.LocalRef, arg interface{}) error {
	m.self = self
	return nil
}

func (m *authForwarder) Started() {
}

func (m *authForwarder) HandleSend(sender actor.Ref, message interface{}) {
	fmt.Println("Handle send unexpected", message)
}

// Will receive "ConnAcceptedAsk" request, should return "ConnAcceptedAnswer" as result.
func (m *authForwarder) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	fmt.Printf("%s receive message, sender:%v type:%T message:%v\n",
		m.self.Id().Name(), sender, ask, ask)
	switch msg := ask.(type) {
	case *websocket.ConnAcceptedAsk:
		return m.newConnAccepted(msg)
	}
	return nil, actor.ErrAskType
}

func (m *authForwarder) newConnAccepted(ask *websocket.ConnAcceptedAsk) (answer *websocket.ConnAcceptedAnswer, err error) {
	fmt.Print("receive new conn, ", ask)
	answer = &websocket.ConnAcceptedAnswer{}
	var name = ask.Header.Get("name")
	var password = ask.Header.Get("password")
	if password != "password" {
		answer.ForbiddenMessageType = websocket.TextMessage
		answer.ForbiddenMessageContent = []byte("Invalid user")
		return answer, errors.New("un-auth user")
	}
	var userRef = actor.ByName(name)
	if userRef == nil {
		userRef, err = actor.SpawnWithName(newUser, name, nil)
		if err != nil {
			answer.ForbiddenMessageType = websocket.TextMessage
			answer.ForbiddenMessageContent = []byte(err.Error())
			return answer, err
		}
	}
	answer.NextForwarder = userRef
	return answer, err
}

func (m *authForwarder) Shutdown() {
}

//
// Listener Actor
//

const listenerName = "Listener"
const listenerAddr = "127.0.0.1:12345"

func newListener() actor.Actor {
	return &websocket.Listener{}
}

//
// User Actor
//

type user struct {
	self *actor.LocalRef
}

func newUser() actor.Actor {
	return &user{}
}

func (m *user) Type() (name string, version int) {
	return "User", 1
}

func (m *user) StartUp(self *actor.LocalRef, arg interface{}) error {
	m.self = self
	return nil
}

func (m *user) Started() {
	fmt.Println("user started")
}

func (m *user) HandleSend(sender actor.Ref, message interface{}) {
	fmt.Printf("user received message:%T\n", message)
	if err := sender.Send(m.self, &websocket.SendText{
		Text: "Received",
	}); err != nil {
		log.Println("send message failed")
		_ = sender.Shutdown(m.self)
	}
}

func (m *user) Shutdown() {
	fmt.Println("user shutdown")
}
