package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/hwangtou/go-actor"
	"log"
	"net/http"
	"time"
)

//
// Actor
//

type Dialer struct {
	self *actor.LocalRef
}

func (m *Dialer) Type() (name string, version int) {
	return "WebsocketDialer", 1
}

func (m *Dialer) StartUp(self *actor.LocalRef, arg interface{}) error {
	m.self = self
	return nil
}

func (m *Dialer) Started() {
}

func (m *Dialer) HandleSend(sender actor.Ref, message interface{}) {
	switch msg := message.(type) {
	case *ReceiveMessage:
		log.Println("websocket dialer receive a message:", string(msg.Buffer), msg.Error)
	case *ReceiveClosed:
		log.Println("websocket dialer receive closed")
	}
}

func (m *Dialer) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	switch msg := ask.(type) {
	// Change forwarding actor
	case *Dialing:
		answer, err = m.dialing(msg)
	default:
		err = actor.ErrMessageValue
	}
	return answer, err
}

func (m *Dialer) Shutdown() {
}

func (m *Dialer) dialing(d *Dialing) (*actor.LocalRef, error) {
	conn, _, err := websocket.DefaultDialer.Dial(d.Url, d.RequestHeader)
	if err != nil {
		log.Println("websocket dialer dialing error,", err)
		return nil, err
	}
	return actor.Spawn(func() actor.Actor { return &connection{} }, &dialConn{
		forwardRef:   d.ForwardRef,
		conn:         conn,
		header:       d.RequestHeader,
		readTimeout:  d.ReadTimeout,
		writeTimeout: d.WriteTimeout,
	})
}

//
// Connect Argument
//

type Dialing struct {
	Url           string
	ForwardRef    actor.Ref
	RequestHeader http.Header
	ReadTimeout   time.Time
	WriteTimeout  time.Time
}
