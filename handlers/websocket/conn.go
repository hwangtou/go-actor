package websocket

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hwangtou/go-actor"
	"log"
	"net/http"
	"time"
)

var (
	ErrForwardActorNotFound   = errors.New("websocket conn actor forward actor not found")
	ErrStartUpUnsupportedType = errors.New("websocket conn start up unsupported type")
)

const (
	TextMessage = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage = websocket.CloseMessage
	PingMessage = websocket.PingMessage
	PongMessage = websocket.PongMessage
)

//
// Actor
//

type connection struct {
	self           *actor.LocalRef
	conn           *websocket.Conn
	reqHeader      http.Header
	readTimeout    time.Time
	writeTimeout   time.Time
	forwarding     actor.Ref
	acceptedOrDial bool
	context        *gin.Context
}

//func HeaderGetter(header map[string][]string, key string) string {
//	return textproto.MIMEHeader(header).Get(key)
//}

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
			m.reqHeader = param.context.Request.Header
			m.readTimeout = param.readTimeout
			m.forwarding = forwarding
			m.acceptedOrDial = true
			m.context = param.context
			return nil
		}
	case *dialConn:
		{
			// get forward actor
			//forwarding := actor.ByName(param.forwardName)
			//if forwarding == nil {
			//	return ErrForwardActorNotFound
			//}
			m.self = self
			m.conn = param.conn
			m.readTimeout = param.readTimeout
			m.forwarding = param.forwardRef
			m.acceptedOrDial = false
			return nil
		}
	default:
		return ErrStartUpUnsupportedType
	}
}

func (m *connection) Started() {
	if !m.acceptedOrDial {
		return
	}
	ask := &ConnAcceptedAsk{
		AcceptedOrDial: m.acceptedOrDial,
		RequestHeaders: m.reqHeader,
	}
	answer := &ConnAcceptedAnswer{}
	if err := m.forwarding.Ask(m.self, ask, &answer); err != nil {
		log.Println("websocket conn started send forwarder error,", err)
		if err := m.self.Shutdown(m.self); err != nil {
			log.Println("websocket conn close error,", err)
		}
		return
	}
	if answer.NextForwarder == nil {
		log.Println("websocket conn started without next forwarder")
		if err := m.self.Shutdown(m.self); err != nil {
			log.Println("websocket conn close error,", err)
		}
		return
	}
	m.changeForwardingActor(answer.NextForwarder)
	go func() {
		for {
			// deadline
			if err := m.conn.SetReadDeadline(m.readTimeout); err != nil {
				log.Println("websocket conn set read deadline error,", err)
				return
			}
			// receive
			msg := &ReceiveMessage{}
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
	case *SendBytes:
		err = m.sendMessage(websocket.BinaryMessage, msg.Buffer)
	case *SendText:
		err = m.sendMessage(websocket.TextMessage, []byte(msg.Text))
	// Handle receive message from connection
	case *ReceiveMessage:
		err = m.forwarding.Send(m.self, msg)
	// Change forwarding actor
	case *ChangeForwardingActor:
		m.changeForwardingActor(msg.NextForwarder)
	default:
		log.Println("websocket conn unsupported send type error, ", msg)
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
	default:
		log.Println("websocket conn unsupported ask type error, ", msg)
		err = actor.ErrMessageValue
	}
	return answer, err
}

func (m *connection) Shutdown() {
	time.AfterFunc(100 * time.Millisecond, func() {
		if err := m.conn.Close(); err != nil {
			log.Println("websocket conn close error,", err)
		}
	})
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
func (m *connection) changeForwardingActor(forwarding actor.Ref) {
	m.forwarding = forwarding
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
	forwardRef actor.Ref
	conn         *websocket.Conn
	readTimeout  time.Time
	writeTimeout time.Time
}

//
// New connection
//

type ConnAcceptedAsk struct {
	AcceptedOrDial bool
	RequestHeaders http.Header
}

type ConnAcceptedAnswer struct {
	NextForwarder actor.Ref
}

//
// Change Forwarding
//

type ChangeForwardingActor struct {
	//ActorName string
	NextForwarder actor.Ref
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

type ReceiveMessage struct {
	MessageType int
	Buffer      []byte
	Error       error
}
