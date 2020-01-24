package chat

import (
	"errors"
	actor "github.com/go-actor"
	"log"
)

type roomManager struct {
	self actor.Ref
	rooms map[string]actor.Ref
}

//
// Actor
//

const NewRoomManagerKey = "room_manager"
const RoomManagerName = "RoomManager"

func NewRoomManager() actor.Actor {
	return &roomManager{
		rooms: map[string]actor.Ref{},
	}
}

func (m *roomManager) StartUp(self actor.Ref, arg interface{}) error {
	log.Println("room manager starting up, args:", arg)
	m.self = self
	return actor.Register(self.Id().ActorId(), "RoomManager")
}

func (m *roomManager) HandleSend(sender actor.Ref, message interface{}) {
	log.Println("room manager handle send, ", sender, message)
}

func (m *roomManager) HandleAsk(sender actor.Ref, message interface{}) (interface{}, error) {
	log.Println("room manager handle ask,", sender, message)
	msg, ok := message.(*roomManagerMessage)
	if !ok {
		return nil, errors.New("room manager handle ask error message type")
	}
	switch msg.action {
	case roomManagerCreate:
		if msg.name == "" {
			return nil, errors.New("room manager handle ask create nil name")
		}
		roomRef, err := actor.Spawn(NewRoomKey, msg.name)
		if err != nil {
			return nil, err
		}
		m.rooms[msg.name] = roomRef
		return &roomMessageReply{}, nil
	case roomManagerAll:
		reply := &roomMessageReply{}
		for name, _ := range m.rooms {
			reply.names = append(reply.names, name)
		}
		return reply, nil
	case roomManagerGet:
		reply := &roomMessageReply{}
		_, reply.hasName = m.rooms[msg.name]
		return reply, nil
	case roomManagerDelete:
		reply := &roomMessageReply{}
		if roomRef, hasName := m.rooms[msg.name]; hasName {
			delete(m.rooms, msg.name)
			roomRef.Release()
			reply.hasName = true
		}
		return reply, nil
	}
	return nil, errors.New("room manager handle ask invalid ask cmd")
}

func (m *roomManager) Idle() {
	log.Println("room manager handle idle")
}

func (m *roomManager) Shutdown() error {
	log.Println("room manager handle shutdown")
	return nil
}

//
// Messages
//

type roomManagerMessage struct {
	action int
	name   string
}

type roomMessageReply struct {
	hasName bool
	names []string
}

const (
	roomManagerCreate = iota
	roomManagerAll
	roomManagerGet
	roomManagerDelete
)
