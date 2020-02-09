package actor

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"reflect"
	"time"
)

var (
	ErrNodeId                = errors.New("actor.Global error node id")
	ErrRemoteRefSendType     = errors.New("actor.Global remote ref send type error")
	ErrRemoteRefAskType     = errors.New("actor.Global remote ref ask type error")
	ErrRemoteRefAnswerType     = errors.New("actor.Global remote ref answer type error")
	ErrGlobalManagerNotReady = errors.New("actor.Global is not ready")
	ErrGlobalNodeNotReady    = errors.New("actor.Global node is not ready")
	ErrRemoteConnNotFound    = errors.New("actor.Global remote conn not found")
	ErrRemoteResponse        = errors.New("actor.Global remote request error")
	ErrRemoteTimeout         = errors.New("actor.Global remote timeout error")
	ErrRemoteActorNotFound   = errors.New("actor.Global remote actor not found")
)

var (
	getNameTimeout = 3 * time.Second
	requestTimeout = 5 * time.Second
)

type globalManager struct {
	sys    *system
	ready  bool
	nodeId uint32
	conn   conn
}

func (m *globalManager) init(sys *system) {
	m.sys = sys
	m.ready = false
}

func (m *globalManager) DefaultInit(nodeId uint32) error {
	if m.ready {
		return ErrGlobalManagerNotReady
	}
	m.conn.global = m
	m.nodeId = nodeId
	if err := m.conn.init(TCP, "0.0.0.0:12345"); err != nil {
		return err
	}
	m.conn.inNames["default"] = ""
	m.ready = true
	return nil
}

func (m *globalManager) Init(nodeId uint32, nw Network, routerAddr string, allowedNames map[string]string) error {
	if m.ready {
		return ErrGlobalManagerNotReady
	}
	m.conn.global = m
	m.nodeId = nodeId
	if err := m.conn.init(nw, routerAddr); err != nil {
		return err
	}
	m.conn.inNames = allowedNames
	m.ready = true
	return nil
}

func (m *globalManager) Close() {
	m.ready = false
	m.nodeId = 0
	m.conn.close()
}

func (m *globalManager) NewConn(nodeId uint32, name, auth string, nw Network, addr string) (*RemoteConn, error) {
	if !m.ready {
		return nil, ErrGlobalManagerNotReady
	}
	c, err := m.conn.getOutConnOrDial(nodeId, name, auth, nw, addr)
	if err != nil {
		return nil, ErrConnError
	}
	return &RemoteConn{
		node: c,
	}, nil
}

func (m *globalManager) GetConn(nodeId uint32, name string) (*RemoteConn, error) {
	if !m.ready {
		return nil, ErrGlobalManagerNotReady
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
	w, err := m.node.send(&ConnControl{
		Type: ControlType_CGetName,
		GetName: &GetName{
			Req: &GetName_Request{
				Name: name,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	select {
	case respMsg, more := <-w.respCh:
		{
			if !more || respMsg == nil || respMsg.getGetNameResp() == nil {
				return nil, ErrRemoteResponse
			}
			resp := respMsg.getGetNameResp()
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
	id Id
	node *outNode  // todo nil
}

func (m RemoteRef) Id() Id {
	return m.id
}

func (m *RemoteRef) Send(sender Ref, msg interface{}) error {
	sendProto, ok := msg.(proto.Message)
	if !ok {
		return ErrRemoteRefAskType
	}

	sendAny, err := ptypes.MarshalAny(sendProto)
	if err != nil {
		return err
	}

	// send request
	w, err := m.node.send(&ConnControl{
		Type:                 ControlType_CSendName,
		SendName: &SendName{
			Req:                  &SendName_Request{
				FromId:   sender.Id().id,
				FromName: sender.Id().name,
				ToName:   m.id.name,
				SendData: sendAny,
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
			if !more || respMsg == nil || respMsg.getSendNameResp() == nil {
				return ErrRemoteResponse
			}
			resp := respMsg.getSendNameResp()
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
	askProto, ok := ask.(proto.Message)
	if !ok {
		return ErrRemoteRefAskType
	}
	answerProto, ok := answer.(proto.Message)
	if !ok {
		return ErrRemoteRefAnswerType
	}

	askAny, err := ptypes.MarshalAny(askProto)
	if err != nil {
		return err
	}

	answerAny, err := ptypes.MarshalAny(answerProto)
	if err != nil {
		return err
	}

	// send request
	w, err := m.node.send(&ConnControl{
		Type:                 ControlType_CAskName,
		AskName:              &AskName{
			Req: &AskName_Request{
				FromId:     sender.Id().id,
				FromName:   sender.Id().name,
				ToName:     m.id.name,
				AskData:    askAny,
				AnswerData: answerAny,
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
			if !more || respMsg == nil || respMsg.getAskNameResp() == nil {
				return ErrRemoteResponse
			}
			resp := respMsg.getAskNameResp()
			var respErr error
			if resp.HasError {
				respErr = errors.New(resp.ErrorMessage)
			}
			err = ptypes.UnmarshalAny(resp.AnswerData, answerProto)
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

func (m *RemoteRef) Shutdown(sender Ref) error {
	// todo
	return errors.New("you should not remote close an actor")
}
