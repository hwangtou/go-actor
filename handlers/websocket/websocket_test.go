package websocket

import (
	"github.com/hwangtou/go-actor"
	"log"
	"testing"
	"time"
)

var listenerRef *actor.LocalRef
var dialerRef *actor.LocalRef

const testAddr = "127.0.0.1:20000"

func TestForwarder(t *testing.T) {
	if _, err := actor.SpawnWithName(func() actor.Actor { return &testForwarder{} }, "forwarder", nil); err != nil {
		log.Println(err)
	}
}

func TestStartListen(t *testing.T) {
	lr, err := actor.SpawnWithName(func() actor.Actor { return &Listener{} }, "WSListener", &StartUpConfig{
		ListenNetwork: TCP4,
		ListenAddr:    testAddr,
		TLS:           nil,
		Handlers: []HandlerConfig{
			{
				Method:           GET,
				RelativePath:     "/websocket",
				ForwardActorName: "forwarder",
				ReadTimeout:      time.Time{},
				WriteTimeout:     time.Time{},
			},
		},
	})
	if err != nil {
		log.Fatalln("test websocket spawn listener error,", err)
	}
	listenerRef = lr
	<-time.After(10 * time.Millisecond)
}

func TestDialer(t *testing.T) {
	dr, err := actor.SpawnWithName(func() actor.Actor { return &Dialer{} }, "WSDialer", nil)
	if err != nil {
		log.Fatalln("test websocket spawn dialer error,", err)
	}
	dialerRef = dr
	<-time.After(10 * time.Millisecond)
}

func TestDialing(t *testing.T) {
	var cr *actor.LocalRef
	if err := dialerRef.Ask(nil, &Dialing{
		Url:         "ws://" + testAddr + "/websocket",
		ForwardName: "forwarder",
	}, &cr); err != nil {
		log.Println(err)
		return
	}
	if err := cr.Send(nil, &SendText{
		Text: "test sending",
	}); err != nil {
		log.Println(err)
		return
	}
	<-time.After(10 * time.Millisecond)
}

func TestSendConn(t *testing.T) {

}

func TestShutdown(t *testing.T) {
	if listenerRef == nil {
		return
	}
	listenerRef.Shutdown(nil)
	dialerRef.Shutdown(nil)
	<-time.After(10 * time.Millisecond)
}

//
// Test Forwarder Actor
//

type testForwarder struct {
	self *actor.LocalRef
}

func (m *testForwarder) Type() (name string, version int) {
	return "testForwarder", 1
}

func (m *testForwarder) StartUp(self *actor.LocalRef, arg interface{}) error {
	m.self = self
	return nil
}

func (m *testForwarder) Started() {
}

func (m *testForwarder) HandleSend(sender actor.Ref, message interface{}) {
	log.Printf("%s receive message, sender:%d type:%T message:%v\n",
		m.self.Id().Name(), sender.Id().ActorId(), message, message)
}

// Will receive "NewWebSocketConnAcceptedAsk" request, should return "NewWebSocketConnAcceptedAnswer" as result.
func (m *testForwarder) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	switch msg := ask.(type) {
	case *NewWebSocketConnAcceptedAsk:
		log.Print("receive new conn, ", msg)
		return &NewWebSocketConnAcceptedAnswer{}, nil
	}
	return nil, actor.ErrAskType
}

func (m *testForwarder) Shutdown() {
}
