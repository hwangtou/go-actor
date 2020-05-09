// Copyright 2020 Tou.Hwang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package actor

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"reflect"
	"time"
)

var (
	//getNameTimeout = 3 * time.Second
	requestTimeout = 5 * time.Second
)

type remoteManager struct {
	sys    *system
	ready  bool
	nodeId uint32
	conn   conn
}

func (m *remoteManager) init(sys *system) {
	m.sys = sys
	m.ready = false
}

func (m *remoteManager) DefaultInit(nodeId uint32) error {
	if m.ready {
		return ErrRemoteManagerNotReady
	}
	m.conn.remote = m
	m.nodeId = nodeId
	if err := m.conn.init(TCP, "0.0.0.0:12345"); err != nil {
		return err
	}
	m.conn.inNames["default"] = ""
	m.ready = true
	return nil
}

func (m *remoteManager) Init(nodeId uint32, nw Network, routerAddr string, allowedNames map[string]string) error {
	if m.ready {
		return ErrRemoteManagerNotReady
	}
	m.conn.remote = m
	m.nodeId = nodeId
	if err := m.conn.init(nw, routerAddr); err != nil {
		return err
	}
	m.conn.inNames = allowedNames
	m.ready = true
	return nil
}

func (m *remoteManager) Close() {
	m.ready = false
	m.nodeId = 0
	m.conn.close()
}

func (m *remoteManager) NewConn(nodeId uint32, name, auth string, nw Network, addr string) (*RemoteConn, error) {
	if !m.ready {
		return nil, ErrRemoteManagerNotReady
	}
	c, err := m.conn.getOutConnOrDial(nodeId, name, auth, nw, addr)
	if err != nil {
		return nil, ErrConnError
	}
	return &RemoteConn{
		node: c,
	}, nil
}

func (m *remoteManager) GetConn(nodeId uint32, name string) (*RemoteConn, error) {
	if !m.ready {
		return nil, ErrRemoteManagerNotReady
	}
	c := m.conn.getOutConn(nodeId, name)
	if c == nil {
		return nil, ErrRemoteConnNotFound
	}
	return &RemoteConn{
		node: c,
	}, nil
}

//
// Remote Node
//

type RemoteConn struct {
	node *outNode
}

func (m *RemoteConn) ByName(name string) (*RemoteRef, error) {
	w, err := m.node.send(&ConnMessage{
		Type: ControlType_CGetName,
		Direction: Direction_Request,
		Content: &ConnMessage_GetName{
			GetName: &GetName{
				Data: &GetName_Req{
					Req: &GetName_Request{
						Name: name,
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	select {
	case respMsg, more := <-w.respCh:
		{
			if !more || respMsg == nil || respMsg.GetGetName() == nil || respMsg.GetGetName().GetResp() == nil {
				return nil, ErrRemoteResponse
			}
			resp := respMsg.GetGetName().GetResp()
			if !resp.Has {
				return nil, ErrRemoteActorNotFound
			}
			return &RemoteRef{
				id: Id{
					node: m.node.nodeId,
					id:   resp.ActorId,
					name: name,
				},
				node: m.node,
			}, nil
		}
	case <-time.After(requestTimeout):
		{
			w.canceled = true
			return nil, ErrRemoteTimeout
		}
	}
}

//
// Remote Ref
//

type RemoteRef struct {
	id   Id
	node *outNode // todo nil
}

func (m RemoteRef) Id() Id {
	return m.id
}

func (m *RemoteRef) Send(sender Ref, msg interface{}) error {
	sendData := &DataContentType{}
	switch obj := msg.(type) {
	case proto.Message:
		sendAny, err := ptypes.MarshalAny(obj)
		if err != nil {
			return err
		}
		sendData.Type = DataType_ProtoBuf
		sendData.Content = &DataContentType_Proto{
			Proto: sendAny,
		}
	case string:
		sendData.Type = DataType_String
		sendData.Content = &DataContentType_Str{
			Str: obj,
		}
	default:
		return ErrRemoteRefAskType
	}

	// send request
	w, err := m.node.send(&ConnMessage{
		Type: ControlType_CSendName,
		Content: &ConnMessage_SendName{
			SendName: &SendName{
				Data: &SendName_Req{
					Req: &SendName_Request{
						FromId:   sender.Id().id,
						FromName: sender.Id().name,
						ToName:   m.id.name,
						SendData: sendData,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// wait for response
	select {
	case respMsg, more := <-w.respCh:
		{
			if !more || respMsg == nil || respMsg.GetSendName() == nil || respMsg.GetSendName().GetResp() == nil {
				return ErrRemoteResponse
			}
			resp := respMsg.GetSendName().GetResp()
			var respErr error
			if resp.HasError {
				respErr = errors.New(resp.ErrorMessage)
			}
			return respErr
		}
	case <-time.After(requestTimeout):
		{
			w.canceled = true
			return ErrRemoteTimeout
		}
	}
}

// todo test answer type not pointer, answer non-struct type, struct contains slice and map
func (m *RemoteRef) Ask(sender Ref, ask interface{}, answer interface{}) error {
	//if t := reflect.TypeOf(answer); t == nil || t.Kind() != reflect.Ptr {
	//	return ErrRemoteRefAnswerType
	//}
	answerValue := reflect.ValueOf(answer)
	if answerValue.Kind() != reflect.Ptr {
		return ErrAnswerType
	}
	answerProto, ok := answer.(proto.Message)
	if !ok {
		return ErrRemoteRefAnswerType
	}

	askData := &DataContentType{}
	switch obj := ask.(type) {
	case proto.Message:
		sendAny, err := ptypes.MarshalAny(obj)
		if err != nil {
			return err
		}
		askData.Type = DataType_ProtoBuf
		askData.Content = &DataContentType_Proto{
			Proto: sendAny,
		}
	case string:
		askData.Type = DataType_String
		askData.Content = &DataContentType_Str{
			Str: obj,
		}
	default:
		return ErrRemoteRefAskType
	}

	answerData := &DataContentType{}
	switch obj := answer.(type) {
	case proto.Message:
		sendAny, err := ptypes.MarshalAny(obj)
		if err != nil {
			return err
		}
		answerData.Type = DataType_ProtoBuf
		answerData.Content = &DataContentType_Proto{
			Proto: sendAny,
		}
	case string:
		answerData.Type = DataType_String
		answerData.Content = &DataContentType_Str{
			Str: obj,
		}
	default:
		return ErrRemoteRefAnswerType
	}

	// send request
	w, err := m.node.send(&ConnMessage{
		Type: ControlType_CAskName,
		Content: &ConnMessage_AskName{
			AskName: &AskName{
				Data: &AskName_Req{
					Req: &AskName_Request{
						FromId:     sender.Id().id,
						FromName:   sender.Id().name,
						ToName:     m.id.name,
						AskData:    askData,
						AnswerData: answerData,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// wait for response
	select {
	case respMsg, more := <-w.respCh:
		{
			if !more || respMsg == nil || respMsg.GetAskName() == nil || respMsg.GetAskName().GetResp() == nil {
				return ErrRemoteResponse
			}
			resp := respMsg.GetAskName().GetResp()
			var respErr error
			if resp.HasError {
				respErr = errors.New(resp.ErrorMessage)
			}
			switch resp.AnswerData.Type {
			case DataType_ProtoBuf:
				err = ptypes.UnmarshalAny(resp.AnswerData.GetProto(), answerProto)
			case DataType_String:
			default:
			}
			if err != nil {
				return err
			}
			if !answerValue.Elem().Type().AssignableTo(reflect.ValueOf(answerProto).Type()) {
				return ErrRemoteRefAnswerType
			}
			answerValue.Elem().Set(reflect.ValueOf(answerProto))
			return respErr
		}
	case <-time.After(requestTimeout):
		{
			w.canceled = true
			return ErrRemoteTimeout
		}
	}
}

// todo should any remote actor send a shutdown message?
func (m *RemoteRef) Shutdown(sender Ref) error {
	// todo
	return errors.New("you should not remote close an actor")
}
