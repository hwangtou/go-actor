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

//
// Remote Node
//

type RemoteConn struct {
	node *outNode
}

func (m *RemoteConn) ByName(name string) (*RemoteRef, error) {
	w, err := m.node.send(&ConnMessage{
		Type:      ControlType_CGetName,
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
	senderId, senderName := uint32(0), ""
	if sender != nil {
		senderId = sender.Id().id
		senderName = sender.Id().name
	}
	w, err := m.node.send(&ConnMessage{
		Type: ControlType_CSendName,
		Content: &ConnMessage_SendName{
			SendName: &SendName{
				Data: &SendName_Req{
					Req: &SendName_Request{
						FromId:   senderId,
						FromName: senderName,
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
	case bool:
		askData.Type = DataType_Bool
		askData.Content = &DataContentType_B{
			B: obj,
		}
	case []byte:
		askData.Type = DataType_Bytes
		askData.Content = &DataContentType_Bs{
			Bs: obj,
		}
	case string:
		askData.Type = DataType_String
		askData.Content = &DataContentType_Str{
			Str: obj,
		}
	}
	w, err := m.node.send(&ConnMessage{
		Type: ControlType_CAskName,
		Content: &ConnMessage_AskName{
			AskName: &AskName{
				Data: &AskName_Req{
					Req: &AskName_Request{
						FromId:     senderId,
						FromName:   senderName,
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
			if resp.AnswerData != nil {
				switch resp.AnswerData.Type {
				case DataType_ProtoBuf:
					answerProto := resp.AnswerData.GetProto()
					err = ptypes.UnmarshalAny(resp.AnswerData.GetProto(), answerProto)
					if err != nil {
						return err
					}
					if !answerValue.Elem().Type().AssignableTo(reflect.ValueOf(answerProto).Type()) {
						return ErrRemoteRefAnswerType
					}
					answerValue.Elem().Set(reflect.ValueOf(answerProto))
				case DataType_String:
					answerString := resp.AnswerData.GetStr()
					if !answerValue.Elem().Type().AssignableTo(reflect.ValueOf(answerString).Type()) {
						return ErrRemoteRefAnswerType
					}
					answerValue.Elem().SetString(answerString)
				default:
				}
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

// todo should any remote actor send a shutdown message?
func (m *RemoteRef) Shutdown(sender Ref) error {
	// todo
	return errors.New("you should not remote close an actor")
}
