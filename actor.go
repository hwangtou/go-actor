// Copyright 2020 Tou.Hwang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*
The goal of go-actor is to make it easier for developers to use the actor model.

Package go-actor provides a portable for creating an actor, with which developer
can receive message that send from actor-system. Actors implements Ask interface
can be asking for, and answer it. Package go-actor initializes a default actor
system, this system help developer to manage actors and their life cycle.

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
*/
package actor

//
// DEVELOPER TO IMPLEMENT
//

// Actor interface
// Implement this interface to spawn it as an actor.
type Actor interface {
	// Return type name and type version of the actor.
	// Name help developer to track an actor.
	// Version start from 1. It is a feature to support hot reload in future.
	Type() (name string, version int)

	// StartUp method and Started method will be called by go-actor, during an actor
	// is starting up. Once an id of an actor has been distributed, StartUp method
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

//
// DEVELOPER TO IMPLEMENT
//

// Actor interface
// Implement this interface to be called by Ask function.
type Ask interface {
	// HandleAsk method will be called to ask an actor to answer a message.
	// It is bidirectional.
	HandleAsk(sender Ref, ask interface{}) (answer interface{}, err error)
}

//
// Go-Actor API
//

var Global *globalManager

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

// Function Type that Create Actor
// Provided by developer.

type Ref interface {
	Id() Id
	Send(sender Ref, msg interface{}) error
	// answer should be a pointer reference to object
	Ask(sender Ref, ask interface{}, answer interface{}) error
	Shutdown(sender Ref) error
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

//
// PRIVATE
//

func init() {
	defaultSys = &system{}
	defaultSys.init()
	Global = &defaultSys.global
}

var defaultSys *system

type system struct {
	locals localsManager
	global globalManager
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
