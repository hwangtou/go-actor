# go-actor

Actor System for Go
The goal of go-actor is to make it easier for developers to use the actor model.

## Get started

Package go-actor provides a portable for creating an actor, with which developer
can receive message that send from actor-system. Actors implements Ask interface
can be asking for, and answer it. Package go-actor initializes a default actor
system, this system help developer to manage actors and their life cycle.

To create an customized actor, you have to implement Actor interface.
```go
type Actor interface {
	// Return type name and type version of the actor.
	// Name help developer to track an actor.
	// Version start from 1. It is a feature to support hot reload in future.
	Type() (name string, version int)

	// StartUp method and Started method will be called by go-actor, during an actor
	// is starting up. Once an id of an actor has been assigned, StartUp method
	// has been call with it's reference, and spawn method argument.
	// StartUp should return an non-nil error, if the StartUp method try to tell
	// go-actor to stop starting an actor, it might cause by invalid argument.
	StartUp(self *LocalRef, arg interface{}) error

	// Started method will be called after StartUp method return nil. This method is
	// to notify that actor is ready.
	Started()

	// HandleSend method will be called to tell an actor that a new message has come.
	// It is unidirectional.
	// WARNING! Do not block this call for a long time. According to the actor model,
	// actor should receive message one by one, to guarantee THREAD-SAFETY. So a long
	// time handling will block all backward messages.
	// Say that again, Since message passing is thread-safe, method HandleSend and
	// HandleAsk will execute one by one. So please do not block this method.
	HandleSend(sender Ref, message interface{})

	// Shutdown method will be called to tell an actor that it will be shutdown, this actor
	// should save data and clean itself up. Actor Reference will shutdown an actor after
	// being call Shutdown, during this process, the Shutdown method of an actor will be
	// called.
	// IMPORTANT: PLEASE SAVE ALL IMPORTANT DATA OF AN ACTOR WHEN IT IS CALLED SHUTDOWN.
	Shutdown()
}
```

Implement this interface to be called by Ask function.

```go
type Ask interface {
	// HandleAsk method will be called to ask an actor to answer a message.
	// It is bidirectional.
	// Answer parameter should be a pointer reference to object.
	HandleAsk(sender Ref, ask interface{}) (answer interface{}, err error)
}
```

Call those spawn functions to create actors and get theirs reference. After
creating an actor instance, developer cannot operate to actor straightly,
but to send message or ask for answer via a reference of an actor. With a
actor reference, developer can also register it a name, and kill an actor.

Inside an actor instance, it will be called StartUp method to try starting up it
with the reference of itself, and the argument that has given in Spawn function.
If an actor started up successfully, its method Started will be called. To receive
others' messages with method HandleSend. And to implement the Ask interface to
support ask-answer pattern. In the end of the life cycle, method Shutdown will be
called, it's the last chance to save data.

Let's take a look at an actor spawning example.

```go
aRef, err := actor.Spawn(func() actor.Actor { return &simpleActor{} }, nil)
if err != nil {
    log.Fatalln(err)
}
if err := actor.Register(aRef, "simple_1"); err != nil {
    log.Fatalln(err)
}
```

Id is a concept that it is an identify of a specific actor. It contains a node id,
a actor id and a actor name. Node id is not zero if the actor is a remote actor,
and this id specify in which remote actor system the actor running. Actor id is
a non-zero number, to help actor system to find it, it just like the ProcessID
of OS. Actor name is a non-blank string if developer register a name to an actor
successfully. Developer should use this name to find an actor, because actors
might be shutdown and startup again, although they are in a same name, but the
actor id must be changed.

You can also use the actor.Global, to connect to other actor.Global, finding
an specific actor with name, send it message, or ask it for an answer. Now global
message should be a ProtoBuf message, ProtoBuf is used in actor.Global serializing.
