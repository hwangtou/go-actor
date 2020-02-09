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
	ErrAnswerType            = errors.New("actor.Local answer type error")
	ErrMessageValue          = errors.New("message value error")
	ErrNodeId                = errors.New("actor.Global error node id")
	ErrRemoteRefSendType     = errors.New("actor.Global remote ref send type error")
	ErrRemoteRefAskType      = errors.New("actor.Global remote ref ask type error")
	ErrRemoteRefAnswerType   = errors.New("actor.Global remote ref answer type error")
	ErrGlobalManagerNotReady = errors.New("actor.Global is not ready")
	ErrGlobalNodeNotReady    = errors.New("actor.Global node is not ready")
	ErrRemoteConnNotFound    = errors.New("actor.Global remote conn not found")
	ErrRemoteResponse        = errors.New("actor.Global remote request error")
	ErrRemoteTimeout         = errors.New("actor.Global remote timeout error")
	ErrRemoteActorNotFound   = errors.New("actor.Global remote actor not found")
	ErrPacketInvalid         = errors.New("conn packet invalid")
	ErrConnError             = errors.New("conn error")
	ErrAuthFailed            = errors.New("conn auth failed")
	ErrAuthTimeout           = errors.New("conn auth timeout")
	ErrReplyFailed           = errors.New("conn reply failed")
)
