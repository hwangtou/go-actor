package go_actor

import (
	"errors"
)

func init() {
	sys.locals.init()
	sys.global.init()
	sys.sessions.init()
}

var (
	ErrActorState         = errors.New("actor state error")
	ErrArgument           = errors.New("argument error")
	ErrNotLocalActor      = errors.New("not local actor")
	ErrActorNotRunning    = errors.New("actor has halt")
	ErrActorCannotAsk     = errors.New("actor cannot ask")
	ErrNameRegistered     = errors.New("name registered")
)

var sys system

type system struct {
	locals   localsManager
	global   globalManager
	sessions sessionsManager
}

//
// Local
//

func Spawn(fn ConstructorFn, arg interface{}) (*LocalRef, error) {
	return sys.locals.spawnActor(fn, "", arg)
}

func SpawnWithName(fn ConstructorFn, name string, arg interface{}) (*LocalRef, error) {
	return sys.locals.spawnActor(fn, name, arg)
}

func Register(ref Ref, name string) error {
	lr, ok := ref.(*LocalRef)
	if !ok {
		return ErrNotLocalActor
	}
	return sys.locals.setNameRunning(lr, name)
}

func ById(id uint32) *LocalRef {
	return sys.locals.getActorRef(id)
}

func ByName(name string) *LocalRef {
	return sys.locals.getName(name)
}

func Watch(id Id) {

}

//func BatchSend(sender Ref, targets []Ref, msg interface{}) (errors map[Id]error) {
//
//}

//
// Global Wrapper
//

var Global globalWrapper

type globalWrapper struct {
}

func (m globalWrapper) NodeId() uint32 {
	return sys.global.nodeId
}

func (m globalWrapper) SpawnWithName(fn ConstructorFn, name string, arg interface{}) (*LocalRef, error) {
	return sys.locals.spawnActor(fn, name, arg)
}

func (m globalWrapper) Register(ref *LocalRef, name string) error {
	return nil
}

func (m globalWrapper) ByName(name string) Ref {
	return nil
}

// ACTOR
// to create an id, construct a struct type that implement Actor and ActorCanAsk interfaces.
// You can not and should not manipulate id directly, because it might destroy atomicity.

// Interface Actor
// It is the base id interface.
type Actor interface {
	Type() (name string, version int)
	StartUp(self *LocalRef, arg interface{}) error
	Started()
	// The method HandleSend, tell means the message is unidirectional.
	// Every id should support this method, to handle basic message passing.
	// Since message passing is thread-safe, method HandleSend and HandleAsk will execute one by one.
	// So please do not block this method if it is not necessary. Consider making it asynchronous.
	// TODO: error return types
	HandleSend(sender Ref, message interface{})

	// Dump()

	// Actor System help you to manage your actors, each id has a reference counter, when the counter
	// decrease to 0, this id will mark as clean up. In this circumstance, method Shutdown will be called.
	// Another circumstance, Actor System support shutting down. This method will be called when it happens.
	// IMPORTANT: PLEASE SAVE ALL IMPORTANT DATA OF AN ACTOR WHEN IT IS CALLED SHUTDOWN.
	Shutdown()
}

type ActorStatus int

const (
	ActorHalt         ActorStatus = 0
	ActorStartingUp   ActorStatus = 1
	ActorRunning      ActorStatus = 2
	ActorShuttingDown ActorStatus = 3
)

type ActorAsk interface {
	HandleAsk(sender Ref, ask interface{}) (answer interface{}, err error)
}

// Function Type that Create Actor
// Provided by developer.
type ConstructorFn = func() Actor

type Ref interface {
	Id() Id
	Send(sender Ref, msg interface{}) error
	Ask(sender Ref, ask interface{}) (answer interface{}, err error)
	Shutdown(sender Ref) error
}

// TODO
//type GroupRef interface {
//	AddRef()
//	AddName()
//	DelRef()
//	DelName()
//	Send()
//	Ask()
//}

//
// Id
//

type Id struct {
	node, id uint32
	name     string
}

func (m Id) NodeId() uint32 {
	return m.node
}

func (m Id) ActorId() uint32 {
	return m.id
}

func (m Id) Name() string {
	return m.name
}
