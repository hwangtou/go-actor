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
const testListener = "testListener"
const testDialer = "testDialer"
const testForwarderName = "forwarder"
const testNextForwarderName = "nextForwarder"

//
// Test
//

func TestForwarder(t *testing.T) {
	log.Println("Spawning forwarder actor")
	if _, err := actor.SpawnWithName(func() actor.Actor { return &testForwarder{} }, testForwarderName, nil); err != nil {
		log.Println(err)
	}
}

func TestNextForwarder(t *testing.T) {
	log.Println("Spawning next forwarder actor")
	if _, err := actor.SpawnWithName(func() actor.Actor { return &testNextForwarder{} }, testNextForwarderName, nil); err != nil {
		log.Println(err)
	}
}

func TestStartListen(t *testing.T) {
	log.Println("Spawning listen actor")
	lr, err := actor.SpawnWithName(func() actor.Actor { return &Listener{} }, testListener, &StartUpConfig{
		IsProduction:  true,
		ListenNetwork: TCP4,
		ListenAddr:    testAddr,
		TLS:           nil,
		Handlers: []HandlerConfig{
			{
				Method:           GET,
				RelativePath:     "/websocket",
				ForwardActorName: testForwarderName,
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
	log.Println("Spawning dialer actor")
	dr, err := actor.SpawnWithName(func() actor.Actor { return &Dialer{} }, testDialer, nil)
	if err != nil {
		log.Fatalln("test websocket spawn dialer error,", err)
	}
	dialerRef = dr
	<-time.After(10 * time.Millisecond)
}

func TestDialing(t *testing.T) {
	log.Println("Dialer actor try dialing")
	var cr *actor.LocalRef
	if err := dialerRef.Ask(nil, &Dialing{
		Url:         "ws://" + testAddr + "/websocket",
		ForwardName: testForwarderName,
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
	log.Panicln("Handle send unexpected")
}

// Will receive "ConnAcceptedAsk" request, should return "ConnAcceptedAnswer" as result.
func (m *testForwarder) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	log.Printf("%s receive message, sender:%d type:%T message:%v\n",
		m.self.Id().Name(), sender.Id().ActorId(), ask, ask)
	switch msg := ask.(type) {
	case *ConnAcceptedAsk:
		log.Print("receive new conn, ", msg)
		return &ConnAcceptedAnswer{
			NextForwarder: actor.ByName(testNextForwarderName),
		}, nil
	}
	return nil, actor.ErrAskType
}

func (m *testForwarder) Shutdown() {
}

//
// Test Next Forwarder
//

type testNextForwarder struct {
	self *actor.LocalRef
}

func (m *testNextForwarder) Type() (name string, version int) {
	return "testNextForwarder", 1
}

func (m *testNextForwarder) StartUp(self *actor.LocalRef, arg interface{}) error {
	m.self = self
	return nil
}

func (m *testNextForwarder) Started() {
}

func (m *testNextForwarder) HandleSend(sender actor.Ref, message interface{}) {
	log.Printf("%s receive message, sender:%d type:%T message:%v\n",
		m.self.Id().Name(), sender.Id().ActorId(), message, message)
	_ = sender.Send(m.self, &SendText{ "response" })
	_ = sender.Shutdown(m.self)
}

func (m *testNextForwarder) Shutdown() {
	panic("Handle ask unexpected")
}
