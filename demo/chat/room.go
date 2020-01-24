package chat

import (
	"errors"
	actor "github.com/go-actor"
	"log"
)

type room struct {
	self actor.Ref
	name string
	joined map[string]actor.Ref
}

//
// Actor
//

const NewRoomKey = "room"

func NewRoom() actor.Actor {
	return &room{
		joined: map[string]actor.Ref{},
	}
}

func (m *room) StartUp(ref actor.Ref, arg interface{}) error {
	log.Println("room starting up, args:", arg)
	roomName, ok := arg.(string)
	if !ok || roomName == "" {
		return errors.New("room starting up arg error")
	}
	if err := actor.Register(ref.Id().ActorId(), roomName); err != nil {
		return err
	}
	m.self = ref
	m.name = roomName
	return nil
}

func (m *room) HandleSend(sender actor.Ref, message interface{}) {
	log.Println("room handle send, ", sender, message)
}

func (m *room) HandleAsk(sender actor.Ref, message interface{}) (interface{}, error) {
	log.Println("room handle ask, ", sender, message)
	msg, ok := message.(*roomMessage)
	if !ok {
		return nil, errors.New("room handle ask error message type")
	}
	switch msg.action {
	case roomJoin:
		m.joined[sender.Id().Name()] = sender
		return &roomReply{}, nil
	case roomLeave:
		delete(m.joined, sender.Id().Name())
		return &roomReply{}, nil
	case roomSend:
		for _, ref := range m.joined {
			if err := ref.Send(m.self, &clientRoomMessage{
				room:     m.name,
				username: msg.name,
				message:  msg.message,
			}); err != nil {
				log.Println("room handle ask send message error,", err)
			}
		}
		return &roomReply{}, nil
	}
	return nil, errors.New("invalid ask cmd")
}

func (m *room) Idle() {
	log.Println("room handle idle", m.name)
}

func (m *room) Shutdown() error {
	log.Println("room handle shutdown", m.name)
	for _, r := range m.joined {
		r.Release()
	}
	m.joined = nil
	return nil
}

//func (m *room) sendMessage(messages ...interface{}) ([]interface{}, error) {
//	s, ok := messages[0].(string)
//	if !ok {
//		return nil, errors.New("message invalid")
//	}
//	for _, r := range m.joined {
//		if err := r.Send(m.self, "new_message", s); err != nil {
//			log.Println(err)
//		}
//	}
//	return []interface{}{"ok"}, nil
//}

//
// Message
//

type roomMessage struct {
	action int
	name string
	message string
}

type roomReply struct {
}

const (
	roomJoin = iota
	roomLeave
	roomSend
)
