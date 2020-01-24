package chat

import (
	"errors"
	"github.com/gin-gonic/gin"
	actor "github.com/go-actor"
	"github.com/gorilla/websocket"
	"log"
	"strings"
)

type client struct {
	self actor.Ref
	conn *websocket.Conn
	connClosed bool
	roomManagerRef actor.Ref
	username string
}

const NewClientKey = "client"

func NewClient() actor.Actor {
	return &client{}
}

func (m *client) StartUp(self actor.Ref, arg interface{}) error {
	ctx, ok := arg.(*gin.Context)
	if !ok {
		return errors.New("client start up with invalid gin.Context")
	}
	upgrade := websocket.Upgrader{}
	ws, err := upgrade.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return err
	}
	m.self = self
	m.conn = ws
	m.roomManagerRef = actor.ByName(RoomManagerName)
	go m.handleConn()
	return nil
}

func (m *client) handleConn() {
	defer func() {
		if !m.connClosed {
			m.connClosed = true
			m.conn.Close()
			m.conn = nil
			m.HandleSend(m.self, &Request{"exit", nil})
		}
	}()
	for {
		req := &Request{}
		err := m.conn.ReadJSON(req)
		if err != nil {
			log.Println("client read error,", err)
			return
		}
		log.Println("client received,", req)
		switch req.Cmd {
		case "exit":
			log.Println("client cmd exit")
			return
		}
		m.HandleSend(m.self, req)
	}
}

func (m *client) send(cmd string, data interface{}, err error) {
	if m.connClosed {
		return
	}
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	if err := m.conn.WriteJSON(Response{
		Cmd:   cmd,
		Data:  data,
		Error: errStr,
	}); err != nil {
		log.Println("client sent error,", err)
	}
}

func (m *client) HandleSend(sender actor.Ref, message interface{}) {
	switch msg := message.(type) {
	case *clientRoomMessage:
		m.send("receive_new_message", msg.message, nil)
	case *Request:
		switch msg.Cmd {
		case "login":
			m.username = msg.Data.(string)
		case "rooms":
			resp, _ := m.roomManagerRef.Ask(m.self, &roomManagerMessage{roomManagerAll, ""})
			rooms := resp.(*roomMessageReply)
			infos := []string{}
			for _, name := range rooms.names {
				infos = append(infos, name)
			}
			m.send("rooms", infos, nil)
		case "create_room":
			m.roomManagerRef.Ask(m.self, &roomManagerMessage{roomManagerCreate, msg.Data.(string)})
			roomRef := actor.ByName(msg.Data.(string))
			if roomRef == nil {
				m.send("create_room", nil, errors.New("room create failed"))
				return
			}
		case "join_room":
			roomRef := actor.ByName(msg.Data.(string))
			if roomRef == nil {
				m.send("send_message", nil, errors.New("room not found"))
				return
			}
			if _, err := roomRef.Ask(m.self, &roomMessage{action:roomJoin}); err != nil {
				m.send("join_room", nil, err)
				return
			}
		case "leave_room":
			roomRef := actor.ByName(msg.Data.(string))
			if roomRef == nil {
				m.send("send_message", nil, errors.New("room not found"))
				return
			}
			if _, err := roomRef.Ask(m.self, &roomMessage{action:roomLeave}); err != nil {
				m.send("leave_room", nil, err)
				return
			}
		case "send_message":
			messages := strings.Split(msg.Data.(string), ">")
			roomRef := actor.ByName(messages[0])
			if roomRef == nil {
				m.send("send_message", nil, errors.New("room not found"))
				return
			}
			_, err := roomRef.Ask(m.self, &roomMessage{
				action:  roomSend,
				name:    m.username,
				message: messages[1],
			})
			if err != nil {
				m.send("send_message", nil, err)
				return
			}
			m.send("send_message", "ok", nil)
		}
	}
}

func (m *client) Idle() {
	log.Println("client handle idle", m.username)
}

func (m *client) Shutdown() error {
	m.roomManagerRef.Release()
	return nil
}

func (m *client) createRoom(name string) {
	_, err := m.roomManagerRef.Ask(m.self, "create_room")
	if err != nil {
		m.send("create_room", nil, err)
		return
	}
	m.send("create_room", nil, nil)
}

//
// Message
//

type clientRoomMessage struct {
	room string
	username string
	message string
}

type Request struct {
	Cmd string
	Data interface{}
}

type Response struct {
	Cmd string
	Data interface{}
	Error string
}
