// Copyright 2020 Tou.Hwang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package actor

import "errors"

var (
	ErrActorState            = errors.New("actor state error")
	ErrArgument              = errors.New("argument error")
	ErrNotLocalActor         = errors.New("not local actor")
	ErrActorNotRunning       = errors.New("actor has halt")
	ErrActorCannotAsk        = errors.New("actor cannot ask")
	ErrNameRegistered        = errors.New("name registered")
	ErrAskType               = errors.New("actor ask type error")
	ErrAnswerType            = errors.New("actor answer type error")
	ErrMessageValue          = errors.New("message value error")
	ErrNodeId                = errors.New("actor.Remote error node id")
	ErrRemoteRefSendType     = errors.New("actor.Remote remote ref send type error")
	ErrRemoteRefAskType      = errors.New("actor.Remote remote ref ask type error")
	ErrRemoteRefAnswerType   = errors.New("actor.Remote remote ref answer type error")
	ErrRemoteManagerNotReady = errors.New("actor.Remote is not ready")
	ErrGlobalNodeNotReady    = errors.New("actor.Remote node is not ready")
	ErrRemoteConnNotFound    = errors.New("actor.Remote remote conn not found")
	ErrRemoteResponse        = errors.New("actor.Remote remote request error")
	ErrRemoteTimeout         = errors.New("actor.Remote remote timeout error")
	ErrRemoteActorNotFound   = errors.New("actor.Remote remote actor not found")
	ErrPacketInvalid         = errors.New("conn packet invalid")
	ErrConnError             = errors.New("conn error")
	ErrAuthFailed            = errors.New("conn auth failed")
	ErrAuthTimeout           = errors.New("conn auth timeout")
	ErrReplyFailed           = errors.New("conn reply failed")
)
