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
// A authForwarder actor for listener actor, is used to handle incoming connection.
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
	log.Println("authForwarder actor started")
}

func (m *authForwarder) HandleSend(sender actor.Ref, message interface{}) {
	fmt.Println("authForwarder actor received send unexpected", message)
}

// Will receive "ConnAcceptedAsk" from listener, should return "ConnAcceptedAnswer" as result.
func (m *authForwarder) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	fmt.Printf("authForwarder actor received ask, sender:%v type:%T message:%v\n",
		sender, ask, ask)
	switch msg := ask.(type) {
	case *websocket.ConnAcceptedAsk:
		return m.newConnAccepted(msg)
	}
	return nil, actor.ErrAskType
}

// Handling new connection accepted, try to get authentication info from request header.
// If user is valid, find the actor of user, and redirect all messages to this user actor.
func (m *authForwarder) newConnAccepted(ask *websocket.ConnAcceptedAsk) (answer *websocket.ConnAcceptedAnswer, err error) {
	fmt.Print("receive new conn, ", ask)
	// Authentication
	answer = &websocket.ConnAcceptedAnswer{}
	var name = ask.Header.Get("name")
	var password = ask.Header.Get("password")
	if password != "password" {
		answer.ForbiddenMessageType = websocket.TextMessage
		answer.ForbiddenMessageContent = []byte("Invalid user")
		return answer, errors.New("un-auth user")
	}
	// Find or create this user actor
	var userRef = actor.ByName(name)
	if userRef == nil {
		userRef, err = actor.SpawnWithName(newUser, name, nil)
		if err != nil {
			answer.ForbiddenMessageType = websocket.TextMessage
			answer.ForbiddenMessageContent = []byte(err.Error())
			return answer, err
		}
	}
	// Redirect to the user actor
	answer.NextForwarder = userRef
	if err := userRef.Send(m.self, &newConnection{}); err != nil {
		return answer, err
	}
	return answer, err
}

func (m *authForwarder) Shutdown() {
	log.Println("authForwarder actor shutdown")
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

type newConnection struct {
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
		fmt.Println("send message failed")
		_ = sender.Shutdown(m.self)
	}
}

func (m *user) Shutdown() {
	fmt.Println("user shutdown")
}
