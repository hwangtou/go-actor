package go_actor

import (
	"errors"
	"log"
)

var (
	ErrNewActorFnNotFound = errors.New("new id function not found")
	ErrActorNameExisted   = errors.New("id name existed")
)

func init() {
	sys.creators = map[string]NewActorFn{}
	sys.manager.init()
	log.Println(sys)
}

var sys system

type system struct {
	creators map[string]NewActorFn
	manager  manager
}

func SetNewActorFn(actorType string, fn NewActorFn) error {
	sys.creators[actorType] = fn
	return nil
}

func UnSetNewActorFn(actorType string) error {
	// TODO
	return nil
}

func Spawn(actorType string) (*Ref, error) {
	fn, has := sys.creators[actorType]
	if !has {
		return nil, ErrNewActorFnNotFound
	}
	a := fn()
	r := sys.manager.addActor(a)
	if err := a.StartUp(r.id); err != nil {
		return nil, err
	}
	return r, nil
}

func Register(id *Id, name string) error {
	return sys.manager.bindName(id.id, name)
}

func ById(id uint32) *Ref {
	r, has := sys.manager.actors[id]
	if !has {
		return nil
	}

	r.countLock.Lock()
	defer r.countLock.Unlock()
	r.count += 1
	return r
}

func ByName(name string) *Ref {
	id, has := sys.manager.names[name]
	if !has {
		return nil
	}
	r, has := sys.manager.actors[id]
	if !has {
		return nil
	}

	r.countLock.Lock()
	defer r.countLock.Unlock()
	r.count += 1
	return r
}

func ShutDown(id uint32) error {
	r, has := sys.manager.actors[id]
	if !has {
		return ErrIdNotFound
	}
	sys.manager.unbindName(r)
	sys.manager.delActor(id)
	r.shutdown()
	return nil
}

// ACTOR
// to create an id, construct a struct type that implement Actor and ActorCanAsk interfaces.
// You can not and should not manipulate id directly, because it might destroy atomicity.

// Interface Actor
// It is the base id interface.
type Actor interface {
	StartUp(id uint32) error
	// The method HandleSend, tell means the message is unidirectional.
	// Every id should support this method, to handle basic message passing.
	// Since message passing is thread-safe, method HandleSend and HandleAsk will execute one by one.
	// So please do not block this method if it is not necessary. Consider making it asynchronous.
	// TODO: error return types
	HandleSend(sender *Id, messages ...interface{})

	Idle()

	// Dump()

	// Actor System help you to manage your actors, each id has a reference counter, when the counter
	// decrease to 0, this id will mark as clean up. In this circumstance, method Shutdown will be called.
	// Another circumstance, Actor System support shutting down. This method will be called when it happens.
	// IMPORTANT: PLEASE SAVE ALL IMPORTANT DATA OF AN ACTOR WHEN IT IS CALLED SHUTDOWN.
	Shutdown() error
}

// Function Type that Create Actor
// Provided by developer.
type NewActorFn = func() Actor
