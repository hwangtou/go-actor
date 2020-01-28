package go_actor

import (
	"errors"
	"log"
)

var (
	ErrNewActorFnNotFound = errors.New("new id function not found")
	ErrActorNameExisted   = errors.New("id nameWrapper existed")
	ErrNotLocalActor = errors.New("not local actor")
)

func init() {
	sys.local.init()
	log.Println(sys)
}

var sys system

type system struct {
	node         uint32
	local        locals
}

func Spawn(fn ConstructorFn, arg interface{}) (*LocalRef, error) {
	return sys.local.spawnActor(fn, "", arg)
}

func SpawnWithName(fn ConstructorFn, name string, arg interface{}) (*LocalRef, error) {
	return sys.local.spawnActor(fn, name, arg)
}

func Register(ref Ref, name string) error {
	lr, ok := ref.(*LocalRef)
	if !ok {
		return ErrNotLocalActor
	}
	return sys.local.setNameRunning(lr, name)
}

func ById(id uint32) *LocalRef {
	return sys.local.getActorRef(id)
}

func ByName(name string) *LocalRef {
	return sys.local.getName(name)
}

// ACTOR
// to create an id, construct a struct type that implement Actor and ActorCanAsk interfaces.
// You can not and should not manipulate id directly, because it might destroy atomicity.

// Interface Actor
// It is the base id interface.
type Actor interface {
	StartUp(self Ref, arg interface{}) error
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
	ActorHalt ActorStatus = 0
	ActorStartingUp ActorStatus = 1
	ActorRunning ActorStatus = 2
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
	IsLocalRunning() bool
	IsRemote() bool
	IsGlobal() bool
	//answer(reply message)
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
