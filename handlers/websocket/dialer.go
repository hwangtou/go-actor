package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/hwangtou/go-actor"
	"log"
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
	log.Println("websocket dialer unknown message from", sender, "message", message)
}

func (m *Dialer) HandleAsk(sender actor.Ref, ask interface{}) (answer interface{}, err error) {
	switch msg := ask.(type) {
	// Change forwarding actor
	case Dialing:
		answer, err = m.dialing(&msg)
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
	conn, _, err := websocket.DefaultDialer.Dial(d.Url, nil)
	if err != nil {
		log.Println("websocket dialer dialing error,", err)
		return nil, err
	}
	return actor.Spawn(func() actor.Actor { return &connection{} }, &dialConn{
		forwardName:  d.ForwardName,
		conn:         conn,
		readTimeout:  d.ReadTimeout,
		writeTimeout: d.WriteTimeout,
	})
}

//
// Connect Argument
//

type Dialing struct {
	Url string
	//RequestHeader http.Header
	ForwardName  string
	ReadTimeout  time.Time
	WriteTimeout time.Time
}
