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

func (m RemoteRef) Status() Status {
	return Halt
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

func interface2ContentType(data interface{}) (c *DataContentType, err error) {
	c = &DataContentType{}
	switch dat := data.(type) {
	case proto.Message:
		sendAny, err := ptypes.MarshalAny(dat)
		if err != nil {
			return
		}
		c.Type = DataType_ProtoBuf
		c.Content = &DataContentType_Proto{
			Proto: sendAny,
		}
		return
	case bool:
		c.Type = DataType_Bool
		c.Content = &DataContentType_B{
			B: dat,
		}
		return
	case []byte:
		c.Type = DataType_Bytes
		c.Content = &DataContentType_Bs{
			Bs: dat,
		}
		return
	case string:
		c.Type = DataType_String
		c.Content = &DataContentType_Str{
			Str: dat,
		}
		return
	case int:
		c.Type = DataType_Int
		c.Content = &DataContentType_I64{
			I64: int64(dat),
		}
		return
	case int8:
		c.Type = DataType_Int8
		c.Content = &DataContentType_I64{
			I64: int64(dat),
		}
		return
	case int16:
		c.Type = DataType_Int16
		c.Content = &DataContentType_I64{
			I64: int64(dat),
		}
		return
	case int32:
		c.Type = DataType_Int32
		c.Content = &DataContentType_I64{
			I64: int64(dat),
		}
		return
	case int64:
		c.Type = DataType_Int64
		c.Content = &DataContentType_I64{
			I64: dat,
		}
		return
	case uint:
		c.Type = DataType_UInt
		c.Content = &DataContentType_U64{
			U64: uint64(dat),
		}
		return
	case uint8:
		c.Type = DataType_UInt8
		c.Content = &DataContentType_U64{
			U64: uint64(dat),
		}
		return
	case uint16:
		c.Type = DataType_UInt16
		c.Content = &DataContentType_U64{
			U64: uint64(dat),
		}
		return
	case uint32:
		c.Type = DataType_UInt32
		c.Content = &DataContentType_U64{
			U64: uint64(dat),
		}
		return
	case uint64:
		c.Type = DataType_UInt64
		c.Content = &DataContentType_U64{
			U64: dat,
		}
		return
	case float32:
		c.Type = DataType_Float32
		c.Content = &DataContentType_F64{
			F64: float64(dat),
		}
		return
	case float64:
		c.Type = DataType_Float64
		c.Content = &DataContentType_F64{
			F64: dat,
		}
		return
	}
	//switch reflect.TypeOf(data).Kind() {
	//case reflect.Slice:
	//case reflect.Map:
	//}
	return c, errors.New("unsupported type")
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

	askData, err := interface2ContentType(ask)
	if err != nil {
		return err
	}
	answerData, err := interface2ContentType(answer)
	if err != nil {
		return err
	}
	senderId, senderName := uint32(0), ""
	if sender != nil {
		senderId, senderName = sender.Id().id, sender.Id().name
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
