package websocket

import (
	"errors"
	"github.com/gin-gonic/gin"
	actor "github.com/go-actor"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

var (
	ErrForwardActorNotFound   = errors.New("websocket conn actor forward actor not found")
	ErrStartUpUnsupportedType = errors.New("websocket conn start up unsupported type")
)

//
// Actor
//

type connection struct {
	self           *actor.LocalRef
	conn           *websocket.Conn
	readTimeout    time.Time
	writeTimeout   time.Time
	forwarding     actor.Ref
	acceptedOrDial bool
}

func (m *connection) Type() (name string, version int) {
	return "WebsocketConn", 1
}

func (m *connection) StartUp(self *actor.LocalRef, arg interface{}) error {
	switch param := arg.(type) {
	case *acceptedConn:
		{
			// get forward actor
			forwarding := actor.ByName(param.forwardName)
			if forwarding == nil {
				return ErrForwardActorNotFound
			}
			// upgrade to websocket
			upgrade := websocket.Upgrader{}
			// TODO header authentication
			ws, err := upgrade.Upgrade(param.context.Writer, param.context.Request, nil)
			if err != nil {
				return err
			}
			m.self = self
			m.conn = ws
			m.readTimeout = param.readTimeout
			m.forwarding = forwarding
			m.acceptedOrDial = true
			return nil
		}
	case *dialConn:
		{
			// get forward actor
			forwarding := actor.ByName(param.forwardName)
			if forwarding == nil {
				return ErrForwardActorNotFound
			}
			m.self = self
			m.conn = param.conn
			m.readTimeout = param.readTimeout
			m.forwarding = forwarding
			m.acceptedOrDial = false
			return nil
		}
	default:
		return ErrStartUpUnsupportedType
	}
}

func (m *connection) Started() {
	if err := m.forwarding.Send(m.self, &NewWebsocketConn{
		AcceptedOrDial: m.acceptedOrDial,
	}); err != nil {
		log.Println("websocket conn started send forwarder error,", err)
		if err := m.self.Shutdown(m.self); err != nil {
			log.Println("websocket conn close error,", err)
		}
		return
	}
	go func() {
		for {
			// deadline
			if err := m.conn.SetReadDeadline(m.readTimeout); err != nil {
				log.Println("websocket conn set read deadline error,", err)
				return
			}
			// receive
			msg := &receiveMessage{}
			msg.MessageType, msg.Buffer, msg.Error = m.conn.ReadMessage()
			if msg.Error != nil {
				// TODO might be block
				if err := m.self.Shutdown(m.self); err != nil {
					log.Println("websocket conn close error,", err)
					return
				}
			}
			// send
			if err := m.self.Send(m.self, msg); err != nil {
				log.Println("websocket conn send error,", err)
				return
			}
		}
	}()
}

func (m *connection) HandleSend(sender actor.Ref, message interface{}) {
	var err error
	switch msg := message.(type) {
	// Handle Send Message from other actor
	case SendBytes:
		err = m.sendMessage(websocket.BinaryMessage, msg.Buffer)
	case *SendBytes:
		err = m.sendMessage(websocket.BinaryMessage, msg.Buffer)
	case SendText:
		err = m.sendMessage(websocket.TextMessage, []byte(msg.Text))
	case *SendText:
		err = m.sendMessage(websocket.TextMessage, []byte(msg.Text))
	// Handle receive message from connection
	case *receiveMessage:
		err = m.forwarding.Send(m.self, msg)
	default:
		err = actor.ErrMessageValue
	}
	if err != nil {
		log.Println("websocket conn handle send error,", err)
		// TODO might be block
		if err := m.self.Shutdown(m.self); err != nil {
			log.Println("websocket conn close error,", err)
		}
	}
}

func (m *connection) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	switch msg := ask.(type) {
	// Change forwarding actor
	case ChangeForwardName:
		err = m.changeForwardingName(msg.ActorName)
	case *ChangeForwardName:
		err = m.changeForwardingName(msg.ActorName)
	default:
		err = actor.ErrMessageValue
	}
	return nil, err
}

func (m *connection) Shutdown() {
	if err := m.conn.Close(); err != nil {
	}
}

//
// Private
//

func (m *connection) sendMessage(t int, buf []byte) error {
	if err := m.conn.SetWriteDeadline(m.writeTimeout); err != nil {
		return err
	}
	if err := m.conn.WriteMessage(t, buf); err != nil {
		return err
	}
	return nil
}

// TODO support remote
func (m *connection) changeForwardingName(name string) error {
	a := actor.ByName(name)
	if a == nil {
		return ErrForwardActorNotFound
	}
	m.forwarding = a
	return nil
}

//
// StartUp Argument
//

type acceptedConn struct {
	forwardName  string
	context      *gin.Context
	readTimeout  time.Time
	writeTimeout time.Time
}

type dialConn struct {
	forwardName  string
	conn         *websocket.Conn
	readTimeout  time.Time
	writeTimeout time.Time
}

//
// New connection
//

type NewWebsocketConn struct {
	AcceptedOrDial bool
}

//
// Change Forwarding
//

type ChangeForwardName struct {
	ActorName string
}

//
// Send Message
//

type SendBytes struct {
	Buffer []byte
}

type SendText struct {
	Text string
}

//
// Read Message
//

type receiveMessage struct {
	MessageType int
	Buffer      []byte
	Error       error
}
