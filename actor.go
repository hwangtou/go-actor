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

//
// DEVELOPER TO IMPLEMENT
//

// Actor interface
// Implement this interface to be called by Ask function.
type Ask interface {
	// HandleAsk method will be called to ask an actor to answer a message.
	// It is bidirectional.
	// Answer parameter should be a pointer reference to object.
	HandleAsk(sender Ref, ask interface{}) (answer interface{}, err error)
}

//
// Go-Actor API
//

// Fast way to get the pointer of globalManager.
var Global *globalManager

// To Spawn an actor locally, with an actor constructor function, and argument to parse
// into the actor StartUp method.
func Spawn(fn func() Actor, arg interface{}) (*LocalRef, error) {
	return defaultSys.Spawn(fn, arg)
}

// Just like what the Spawn function do. And there is an name parameter, to register what
// name the local actor is. This spawn function will try to register the name first, if
// failed, it will return ErrNameRegistered.
// This spawn function is tread-safe, it guarantee when StartUp method has been called,
// the name is already bound to the local actor.
func SpawnWithName(fn func() Actor, name string, arg interface{}) (*LocalRef, error) {
	return defaultSys.SpawnWithName(fn, name, arg)
}

// Register function, it try to bind a name to a local actor via its reference.
// Registered name is unique in an actor-system, developer cannot registered the same
// name, until the name has been un-registered.
func Register(ref *LocalRef, name string) error {
	return defaultSys.Register(ref, name)
}

// Get a local reference with its actor id.
// Not Recommended to use
func ById(id uint32) *LocalRef {
	return defaultSys.ById(id)
}

// Once an local actor has registered to a name, we can get its reference by name.
func ByName(name string) *LocalRef {
	return defaultSys.ByName(name)
}

// Ref is short for reference.
// It's a kind of instances that interact with the real actor instance.
// There are two type of Ref: *LocalRef and *RemoteRef.
// LocalRef means that the real actor is running locally in the same actor-system.
// RemoteRef means that the real actor is running remotely.
type Ref interface {
	// Get Id of an actor
	Id() Id
	// Send message to actor via reference. HandleSend method of the actor will be called.
	Send(sender Ref, msg interface{}) error
	// Ask an actor via reference. HandleAsk method of the actor will be called.
	// Answer parameter should be a pointer reference to object.
	Ask(sender Ref, ask interface{}, answer interface{}) error
	// Shutdown an actor via reference. Shutdown method of the actor will be called.
	// todo shutdown cause by panic
	Shutdown(sender Ref) error
}

// Id is short for identify.
// It's a struct that contains node name, actor id and actor name of an actor.
type Id struct {
	node, id uint32
	name     string
}

// Node id is 0, means that actor running locally.
// Node id is greater than 0, means that actor running in a remote actor system,
// the positive number is the id of the actor system.
func (m Id) NodeId() uint32 {
	return m.node
}

// Actor id is greater than 0, otherwise the id is invalid.
// Actor id is an specific positive number, which is assigned by actor-system when
// an actor is being allocating, is increasing from 1 and auto increase.
// Actor id help actor system to find it, it just like the ProcessID of OS.
func (m Id) ActorId() uint32 {
	return m.id
}

// Name is non-blank, means that the actor has been already bound to a name.
// Use ByName function to get the reference of the actor, if the actor still alive.
func (m Id) Name() string {
	return m.name
}

//
// PRIVATE
//

func init() {
	defaultSys = NewSystem()
	Global = &defaultSys.global
}

// Go-actor provides a default system instance for use.
// Developer can create system instance if needed, but not recommended.
var defaultSys *system

// It's the core of go-actor.
type system struct {
	locals localsManager
	global globalManager
}

// Developer can create system instance if needed, but not recommended.
func NewSystem() *system {
	defaultSys = &system{}
	defaultSys.init()
	return defaultSys
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

// todo: to watch an id, notify when the actor of this id has been shutdown.
//func Watch(id Id) {
//
//}

// todo: to batch send a message to a basket of reference, no matter LocalRef or RemoteRef
// todo: group send
//func BatchSend(sender Ref, targets []Ref, msg interface{}) (errors map[Id]error) {
//
//}
//type GroupRef interface {
//	AddRef()
//	AddName()
//	DelRef()
//	DelName()
//	Send()
//	Ask()
//}

// todo: performance monitoring system
// running actors
// timeout session
