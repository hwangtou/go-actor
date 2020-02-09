package actor

import (
	"errors"
)


var (
	ErrActorState         = errors.New("actor state error")
	ErrArgument           = errors.New("argument error")
	ErrNotLocalActor      = errors.New("not local actor")
	ErrActorNotRunning    = errors.New("actor has halt")
	ErrActorCannotAsk     = errors.New("actor cannot ask")
	ErrNameRegistered     = errors.New("name registered")
)


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

type Status int

const (
	Halt         Status = 0
	StartingUp   Status = 1
	Running      Status = 2
	ShuttingDown Status = 3
)

type Ask interface {
	HandleAsk(sender Ref, ask interface{}) (answer interface{}, err error)
}

// Function Type that Create Actor
// Provided by developer.

type Ref interface {
	Id() Id
	Send(sender Ref, msg interface{}) error
	// answer should be a pointer reference to object
	Ask(sender Ref, ask interface{}, answer interface{}) error
	Shutdown(sender Ref) error
}


func init() {
	defaultSys = &system{}
	defaultSys.init()
	Global = &defaultSys.global
}

var defaultSys *system
var Global *globalManager


type system struct {
	locals   localsManager
	global   globalManager
}

func (m *system) init() {
	m.locals.init(m)
	m.global.init(m)
}

func (m *system) Spawn(fn func() Actor, arg interface{}) (*LocalRef, error) {
	return m.locals.spawnActor(fn, "", arg)
}

func (m *system) SpawnWithName(fn func() Actor, name string, arg interface{}) (*LocalRef, error) {
	return m.locals.spawnActor(fn, name, arg)
}

func (m *system) Register(ref Ref, name string) error {
	lr, ok := ref.(*LocalRef)
	if !ok {
		return ErrNotLocalActor
	}
	return m.locals.setNameRunning(lr, name)
}

func (m *system) ById(id uint32) *LocalRef {
	return m.locals.getActorRef(id)
}

func (m *system) ByName(name string) *LocalRef {
	return m.locals.getName(name)
}

func (m *system) Global() *globalManager {
	return &m.global
}


func Spawn(fn func() Actor, arg interface{}) (*LocalRef, error) {
	return defaultSys.Spawn(fn, arg)
}

func SpawnWithName(fn func() Actor, name string, arg interface{}) (*LocalRef, error) {
	return defaultSys.SpawnWithName(fn, name, arg)
}

func Register(ref Ref, name string) error {
	return defaultSys.Register(ref, name)
}

func ById(id uint32) *LocalRef {
	return defaultSys.ById(id)
}

func ByName(name string) *LocalRef {
	return defaultSys.ByName(name)
}

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

// todo
//func Watch(id Id) {
//
//}

// todo
//func BatchSend(sender Ref, targets []Ref, msg interface{}) (errors map[Id]error) {
//
//}

// todo group send
//type GroupRef interface {
//	AddRef()
//	AddName()
//	DelRef()
//	DelName()
//	Send()
//	Ask()
//}

// todo performance
// running actors
// timeout session
