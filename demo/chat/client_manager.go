package chat

import (
	"github.com/gin-gonic/gin"
	actor "github.com/go-actor"
	"log"
)

type clientManager struct {
	self actor.Ref
	ws *gin.Engine
	clients map[*actor.Id]actor.Ref
}

const NewClientManagerKey = "client_manager"

func NewClientManager() actor.Actor {
	return &clientManager{
		ws: gin.Default(),
		clients: map[*actor.Id]actor.Ref{},
	}
}

func (m *clientManager) StartUp(self actor.Ref, arg interface{}) error {
	log.Println("client manager starting up, args:", arg)
	m.self = self
	//listener, err := net.Listen("tcp", args[0].(string))
	//if err != nil {
	//	return err
	//}
	m.ws.GET("/chat", func(context *gin.Context) {
		m.HandleSend(m.self, &clientManagerMessage{"new_client", context})
	})
	go m.ws.Run(arg.(string))
	return nil
}

func (m *clientManager) HandleSend(sender actor.Ref, message interface{}) {
	msg, ok := message.(*clientManagerMessage)
	if !ok {
		log.Println("client manager handle send invalid message")
		return
	}
	switch msg.cmd {
	case "new_client":
		m.newClient(msg.conn)
	}
}

func (m *clientManager) Idle() {
	panic("implement me")
}

func (m *clientManager) Shutdown() error {
	panic("implement me")
}

func (m *clientManager) newClient(context *gin.Context) {
	clientRef, err := actor.Spawn(NewClientKey, context)
	if err != nil {
		log.Println("client spawn failed,", err)
		return
	}
	m.clients[clientRef.Id()] = clientRef
}

type clientManagerMessage struct {
	cmd string
	conn *gin.Context
}
