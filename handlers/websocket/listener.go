package websocket

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hwangtou/go-actor"
	"log"
	"net"
	"net/http"
	"time"
)

var (
	ErrStarUpConfig = errors.New("websocket actor start up config error")
)

type HttpMethod string

const (
	GET  HttpMethod = "GET"
	POST HttpMethod = "POST"
)

type Network string

const (
	TCP  Network = "tcp"
	TCP4 Network = "tcp4"
	TCP6 Network = "tcp6"
)

const (
	shutdownTimeout = 5 * time.Second
)

//
// Actor
//

type Listener struct {
	self     *actor.LocalRef
	config   StartUpConfig
	listener net.Listener
	server   *http.Server
}

func (m *Listener) Type() (name string, version int) {
	return "WebsocketListener", 1
}

func (m *Listener) StartUp(self *actor.LocalRef, arg interface{}) (err error) {
	// check config
	cfg, ok := arg.(*StartUpConfig)
	if !ok {
		return ErrStarUpConfig
	}
	if len(cfg.Handlers) == 0 || cfg.ListenAddr == "" {
		return ErrStarUpConfig
	}
	// set router
	h := gin.Default()
	for _, handler := range cfg.Handlers {
		h.Handle(string(handler.Method), handler.RelativePath, func(context *gin.Context) {
			if _, err := actor.Spawn(func() actor.Actor { return &connection{} }, &acceptedConn{
				forwardName:  handler.ForwardActorName,
				context:      context,
				readTimeout:  handler.ReadTimeout,
				writeTimeout: handler.WriteTimeout,
			}); err != nil {
				log.Println("websocket listener new conn actor spawn error,", err)
			}
		})
	}
	if cfg.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	// server
	var l net.Listener
	if cfg.TLS != nil {
		l, err = tls.Listen(string(cfg.ListenNetwork), cfg.ListenAddr, cfg.TLS)
	} else {
		l, err = net.Listen(string(cfg.ListenNetwork), cfg.ListenAddr)
	}
	if err != nil {
		return err
	}

	m.self = self
	m.config = *cfg
	m.listener = l
	m.server = &http.Server{
		Handler: h,
	}
	return nil
}

func (m *Listener) Started() {
	go func() {
		if err := m.server.Serve(m.listener); err != nil {
			log.Println("websocket listener listen error,", err)
		}
	}()
}

func (m *Listener) HandleSend(sender actor.Ref, message interface{}) {
	log.Println("websocket listener unknown message from", sender, "message", message)
}

func (m *Listener) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := m.server.Shutdown(ctx); err != nil {
		log.Println("websocket listener server shutdown error,", err)
	}
}

//
// StartUp Argument
//

type StartUpConfig struct {
	IsProduction  bool
	ListenNetwork Network
	ListenAddr    string
	TLS           *tls.Config
	Handlers      []HandlerConfig
}

type HandlerConfig struct {
	Method           HttpMethod
	RelativePath     string
	ForwardActorName string
	ReadTimeout      time.Time
	WriteTimeout     time.Time
}
