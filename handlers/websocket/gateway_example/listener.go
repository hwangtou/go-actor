package main

import (
	"errors"
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
	return "authForwarder", 1
}

func (m *authForwarder) StartUp(self *actor.LocalRef, arg interface{}) error {
	m.self = self
	return nil
}

func (m *authForwarder) Started() {
}

func (m *authForwarder) HandleSend(sender actor.Ref, message interface{}) {
	log.Panicln("Handle send unexpected")
}

// Will receive "ConnAcceptedAsk" request, should return "ConnAcceptedAnswer" as result.
func (m *authForwarder) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	log.Printf("%s receive message, sender:%d type:%T message:%v\n",
		m.self.Id().Name(), sender.Id().ActorId(), ask, ask)
	switch msg := ask.(type) {
	case *websocket.ConnAcceptedAsk:
		return m.newConnAccepted(msg)
	}
	return nil, actor.ErrAskType
}

func (m *authForwarder) newConnAccepted(ask *websocket.ConnAcceptedAsk) (answer *websocket.ConnAcceptedAnswer, err error) {
	log.Print("receive new conn, ", ask)
	var name = ask.RequestHeaders.Get("name")
	var password = ask.RequestHeaders.Get("password")
	if password != "password" {
		return answer, errors.New("un-auth user")
	}
	var userRef = actor.ByName(name)
	if userRef == nil {
		userRef, err = actor.SpawnWithName(newUser, name, nil)
		if err != nil {
			return answer, err
		}
	}
	answer = &websocket.ConnAcceptedAnswer{
		NextForwarder: userRef,
	}
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
	return "user", 1
}

func (m *user) StartUp(self *actor.LocalRef, arg interface{}) error {
	m.self = self
	return nil
}

func (m *user) Started() {
}

func (m *user) HandleSend(sender actor.Ref, message interface{}) {
}

func (m *user) Shutdown() {
}
