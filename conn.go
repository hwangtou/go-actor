// Copyright 2020 Tou.Hwang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package actor

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

//
// conn
//

type Network string

const (
	TCP  Network = "tcp"
	TCP4 Network = "tcp4"
	TCP6 Network = "tcp6"
)

const (
	readTimeout            = 0
	readBufferSize         = 1024
	packetSizeLimit        = 1024 * 1024
	packetHeaderSize       = 4 // Int
	packetChannelSize      = 5
	authTimeout            = 5
	inMessageChannelLength = 10
)

type conn struct {
	global      *globalManager
	ready       bool
	listener    net.Listener
	inNames     map[string]string
	inConn      map[uint32]map[string]*inNode
	inConnLock  sync.RWMutex
	inMessageCh chan *inReply
	outConn     map[uint32]map[string]*outNode
	outConnLock sync.RWMutex
}

func (m *conn) init(nw Network, listen string) error {
	l, err := net.Listen(string(nw), listen)
	if err != nil {
		m.ready = false
		log.Println("actor.Global.init error,", err)
		return err
	}
	m.ready = true
	m.listener = l
	m.inNames = make(map[string]string)
	m.inConn = make(map[uint32]map[string]*inNode)
	m.inMessageCh = make(chan *inReply, inMessageChannelLength)
	m.outConn = make(map[uint32]map[string]*outNode)
	go m.inMessageHandle()
	go m.loop()
	return nil
}

func (m *conn) inMessageHandle() {
	log.Println("start in message loop")
	for {
		msg, more := <-m.inMessageCh
		if !more {
			break
		}
		if msg == nil || msg.inConn == nil || msg.inMessage == nil ||
			msg.inMessage.Control == nil {
			log.Println("actor.Global in message handle error,", msg)
			continue
		}
		go func() {
			defer func() {
				// return
			}()
			switch msg.inMessage.getType() {
			case ControlType_CSendName:
				{
					sendName := msg.inMessage.getSendNameReq()
					if sendName == nil {
						log.Println("actor.Global in message control send name error,", msg)
						return
					}
					lr := m.global.sys.locals.getName(sendName.ToName)
					if lr == nil {
						return
					}
					var (
						fromRef *RemoteRef
					)
					if sendName.FromId != 0 {
						fromRef = &RemoteRef{
							id: Id{
								node: msg.inConn.nodeId,
								id:   sendName.FromId,
								name: sendName.FromName,
							},
							node: nil,
						}
					}
					sendD, err := ptypes.Empty(sendName.SendData)
					if err != nil {
						return
					}
					err = ptypes.UnmarshalAny(sendName.SendData, sendD)
					if err != nil {
						return
					}
					err = lr.Send(fromRef, sendD)
					if err != nil {
						return
					}
				}
			case ControlType_CAskName:
				{
					askName := msg.inMessage.getAskNameReq()
					if askName == nil {
						log.Println("actor.Global in message control ask name error,", msg)
						return
					}
					lr := m.global.sys.locals.getName(askName.ToName)
					if lr == nil {
						return
					}
					var (
						fromRef *RemoteRef
					)
					if askName.FromId != 0 {
						fromRef = &RemoteRef{
							id: Id{
								node: msg.inConn.nodeId,
								id:   askName.FromId,
								name: askName.FromName,
							},
							node: nil,
						}
					}
					askD, err := ptypes.Empty(askName.AskData)
					if err != nil {
						return
					}
					answerD, err := ptypes.Empty(askName.AnswerData)
					if err != nil {
						return
					}
					err = ptypes.UnmarshalAny(askName.AskData, askD)
					if err != nil {
						return
					}
					err = lr.Ask(fromRef, askD, &answerD)
					if err != nil {
						return
					}
					answerAny, err := ptypes.MarshalAny(answerD)
					if err != nil {
						return
					}
					err = msg.reply(&ConnControl{
						Type: ControlType_CAskName,
						AskName: &AskName{
							Resp: &AskName_Response{
								HasError:     false,
								ErrorMessage: "",
								AnswerData:   answerAny,
							},
						},
					})
					if err != nil {
						return
					}
				}
			case ControlType_CGetName:
				{
					getName := msg.inMessage.getGetNameReq()
					if getName == nil {
						log.Println("actor.Global in message control get name error,", msg)
						return
					}
					var (
						hasActor bool   = false
						actorId  uint32 = 0
					)
					if lr := m.global.sys.locals.getName(getName.Name); lr != nil {
						hasActor = true
						actorId = lr.id.id
					}
					if err := msg.reply(&ConnControl{
						Type: ControlType_CGetName,
						GetName: &GetName{
							Resp: &GetName_Response{
								Has:     hasActor,
								ActorId: actorId,
							},
						},
					}); err != nil {
						log.Println("actor.Global in message control get name reply error,", err)
					}
				}
			default:
				log.Println("actor.Global in message control type error,", msg)
			}
		}()
	}
}

func (m *conn) loop() {
	log.Println("start listen loop")
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			break
		}
		go func() {
			n := &inNode{
				ready:  false,
				global: m.global,
				nodeId: 0,
				name:   "",
				conn: connSafe{
					Conn:   conn,
					Mutex:  sync.Mutex{},
					closed: false,
				},
				reader: connReader{},
				writer: connWriter{},
			}
			n.reader.init(&n.conn)
			n.writer.init(&n.conn)

			// todo connect back
			defer func() {
				n.conn.safeClose()
			}()
			// receive auth
			select {
			case packet, more := <-n.reader.recvCh:
				{
					log.Println(">>>>>>>>>>conn auth packet,", packet)
					reply := func(isAuth bool) {
						n.writer.send(&ConnMessage{
							SequenceId: packet.SequenceId,
							Control: &ConnControl{
								Type: ControlType_CAuth,
								Auth: &Auth{
									Resp: &Auth_Response{
										IsAuth: isAuth,
									},
								},
							},
						})
					}
					if !more {
						log.Println("actor.Global.listen auth no message", packet)
						// todo reply error
						reply(false)
						return
					}
					if packet == nil || packet.getAuthReq() == nil {
						log.Println("actor.Global.listen auth info error", packet)
						// todo reply error
						reply(false)
						return
					}
					password, has := m.inNames[packet.Control.Auth.Req.ConnName]
					if !has || password != packet.Control.Auth.Req.Password {
						log.Println("actor.Global.listen auth invalid,", packet)
						// todo reply error
						reply(false)
						return
					}
					if packet.Control.Auth.Req.FromNodeId == 0 {
						log.Println("actor.Global.listen auth invalid,", packet)
						// todo reply error
						reply(false)
						return
					}
					if packet.Control.Auth.Req.ToNodeId != m.global.nodeId {
						log.Println("actor.Global.listen auth invalid,", packet)
						// todo reply error
						reply(false)
						return
					}
					n.nodeId = packet.Control.Auth.Req.FromNodeId
					n.name = packet.Control.Auth.Req.ConnName
					if _, has := m.inConn[n.nodeId]; !has {
						m.inConn[n.nodeId] = make(map[string]*inNode)
					}
					if oldN, has := m.inConn[n.nodeId][n.name]; has {
						oldN.conn.safeClose()
					}
					m.inConn[n.nodeId][n.name] = n
					// todo reply okay
					reply(true)

					// todo binary dial
				}
			case <-time.After(authTimeout * time.Second):
				log.Println("actor.Global.listen auth timeout")
				return
			}
			// reply
			for {
				log.Println(">>>>>>>>>>conn loop")
				packet, more := <-n.reader.recvCh
				log.Println(">>>>>>>>>>conn loop packet,", packet)
				if !more {
					return
				}
				m.inMessageCh <- &inReply{
					inMessage: packet,
					inConn:    n,
				}
			}
		}()
	}
}

func (m *conn) getOutConn(nodeId uint32, name string) *outNode {
	if nodeId == 0 {
		return nil
	}

	// get conn
	m.outConnLock.RLock()
	if _, has := m.outConn[nodeId]; !has {
		m.outConn[nodeId] = make(map[string]*outNode)
	}
	n, _ := m.outConn[nodeId][name]
	m.outConnLock.RUnlock()
	return n
}

func (m *conn) getOutConnOrDial(nodeId uint32, name, auth string, nw Network, addr string) (*outNode, error) {
	if nodeId == 0 {
		return nil, ErrNodeId
	}

	m.outConnLock.Lock()
	defer m.outConnLock.Unlock()

	// get conn
	if _, has := m.outConn[nodeId]; !has {
		m.outConn[nodeId] = make(map[string]*outNode)
	}
	n, has := m.outConn[nodeId][name]
	if has {
		return n, nil
	}

	// todo
	n = &outNode{
		ready:   false,
		global:  m.global,
		nodeId:  nodeId,
		name:    name,
		nw:      nw,
		addr:    addr,
		conn:    connSafe{},
		reader:  connReader{},
		writer:  connWriter{},
		seq:     make(map[uint64]*seqWrapper),
		seqId:   0,
		seqLock: sync.Mutex{},
	}
	m.outConn[nodeId][name] = n
	if err := n.dial(name, auth); err != nil {
		delete(m.outConn[nodeId], name)
		return nil, err
	}
	log.Println("actor.Global.getOutConnOrDial new conn")
	return n, nil
}

func (m *conn) close() {
	m.ready = false
	m.listener.Close()
	m.inNames = map[string]string{}
	m.inConnLock.Lock()
	for nodeId, nodes := range m.inConn {
		for name, node := range nodes {
			node.conn.safeClose()
			delete(nodes, name)
		}
		delete(m.inConn, nodeId)
	}
	m.inConnLock.Unlock()
	m.outConnLock.Lock()
	for nodeId, nodes := range m.outConn {
		for name, node := range nodes {
			node.conn.safeClose()
			for id, seq := range node.seq {
				seq.respCh <- nil
				delete(node.seq, id)
			}
			delete(nodes, name)
		}
		delete(m.outConn, nodeId)
	}
	m.outConnLock.Unlock()
}

//
// out conn node
//

type outNode struct {
	ready   bool
	global  *globalManager
	nodeId  uint32
	name    string
	nw      Network
	addr    string
	conn    connSafe
	reader  connReader
	writer  connWriter
	seq     map[uint64]*seqWrapper
	seqId   uint64
	seqLock sync.Mutex
}

type seqWrapper struct {
	req       *ConnMessage
	respCh    chan *ConnMessage
	canceled  bool
	createdAt time.Time
}

func (m *outNode) dial(name, password string) (err error) {
	m.ready = false
	m.conn.closed = false
	m.conn.Conn, err = net.Dial(string(m.nw), m.addr)
	if err != nil {
		log.Println("actor.Global.getOutConnOrDial dial error,", err)
		return err
	}
	m.reader.init(&m.conn)
	m.writer.init(&m.conn)

	err = m.auth(m.nodeId, name, password)
	if err != nil {
		log.Println("actor.Global.getOutConnOrDial password error,", err)
		m.conn.safeClose()
		return err
	}

	m.ready = true
	go m.loop()
	return nil
}

func (m *outNode) auth(nodeId uint32, name, password string) error {
	// send authentication
	if err := m.writer.send(&ConnMessage{
		SequenceId: 0,
		Control: &ConnControl{
			Type: ControlType_CAuth,
			Auth: &Auth{
				Req: &Auth_Request{
					FromNodeId: m.global.nodeId,
					ToNodeId:   nodeId,
					ConnName:   name,
					Password:   password,
				},
			},
		},
	}); err != nil {
		return err
	}
	// receive
	select {
	case packet, more := <-m.reader.recvCh:
		log.Println(">>>>>>>>>>out node auth packet,", packet)
		if !more {
			return ErrConnError
		}
		if packet == nil || packet.getAuthResp() == nil {
			return ErrAuthFailed
		}
		if packet.Control.Auth.Resp.IsAuth {
			return nil
		}
		log.Println("actor.Global.getOutConnOrDial auth error message", packet.Control)
		return ErrAuthFailed
	case <-time.After(authTimeout * time.Second):
		return ErrAuthTimeout
	}
}

func (m *outNode) loop() {
	for {
		packet, more := <-m.reader.recvCh
		log.Println(">>>>>>>>>>out node loop packet,", packet)
		if !more {
			return
		}
		seq, has := m.seq[packet.SequenceId]
		if !has {
			log.Println("actor.Global out node receive unknown sequence,", packet)
			continue
		}
		m.seqLock.Lock()
		delete(m.seq, packet.SequenceId)
		m.seqLock.Unlock()
		seq.respCh <- packet
	}
}

func (m *outNode) send(control *ConnControl) (*seqWrapper, error) {
	if !m.ready {
		return nil, ErrGlobalNodeNotReady
	}

	m.seqLock.Lock()
	for {
		m.seqId++
		if _, has := m.seq[m.seqId]; !has {
			break
		}
	}
	seqId := m.seqId
	w := &seqWrapper{
		req: &ConnMessage{
			SequenceId: seqId,
			Control:    control,
		},
		respCh:    make(chan *ConnMessage, 1),
		createdAt: time.Now(),
	}
	m.seq[seqId] = w
	m.seqLock.Unlock()

	if err := m.writer.send(w.req); err != nil {
		m.seqLock.Lock()
		delete(m.seq, seqId)
		m.seqLock.Unlock()
		return nil, err
	}
	return w, nil
}

func (m *outNode) close() {
	m.ready = false
	m.conn.safeClose()
	m.seqLock.Lock()
	for id, w := range m.seq {
		w.respCh <- nil
		delete(m.seq, id)
	}
	m.seqLock.Unlock()
}

//
// in conn node
//

type inNode struct {
	ready  bool
	global *globalManager
	nodeId uint32
	name   string
	conn   connSafe
	reader connReader
	writer connWriter
}

type inReply struct {
	inMessage *ConnMessage
	inConn    *inNode
}

func (m *inReply) reply(control *ConnControl) error {
	if !m.inConn.ready {
		return ErrReplyFailed
	}
	return m.inConn.writer.send(&ConnMessage{
		SequenceId: m.inMessage.SequenceId,
		Control:    control,
	})
}

//
// conn reader
//

type connReader struct {
	conn     *connSafe
	recvCh   chan *ConnMessage
	inBuffer []byte
	header   []byte
	body     []byte
	size     int
}

func (m *connReader) init(conn *connSafe) {
	m.conn = conn
	m.recvCh = make(chan *ConnMessage, packetChannelSize)
	m.inBuffer = make([]byte, 0, packetSizeLimit)
	m.header = make([]byte, 0, packetHeaderSize)
	m.body = nil
	m.size = 0
	go m.loop()
}

func (m *connReader) loop() {
	buf := make([]byte, readBufferSize)
	func() {
		defer func() {
			m.conn.safeClose()
			close(m.recvCh)
		}()
		for {
			if readTimeout > 0 {
				if err := m.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
					return
				}
			}
			num, err := m.conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println("actor.Global out conn loop error,", err)
				}
				return
			}
			buffers, err := m.handleBuffer(buf[:num])
			for _, buf := range buffers {
				p := &ConnMessage{}
				if err := proto.Unmarshal(buf, p); err != nil {
					return
				}
				m.recvCh <- p
			}
			if err != nil {
				return
			}
		}
	}()
}

func (m *connReader) handleBuffer(buf []byte) (buffers [][]byte, err error) {
	// Copy data to prevent data modifying
	copiedBuf := make([]byte, len(buf))
	copy(copiedBuf, buf)
	m.inBuffer = append(m.inBuffer, copiedBuf...)
	buffers = [][]byte{}
	for {
		if len(m.inBuffer) == 0 {
			break
		}
		has, err := func() (bool, error) {
			if headerNeed := packetHeaderSize - len(m.header); headerNeed > 0 {
				// Get header
				switch {
				case len(m.inBuffer) == 0:
					return false, nil
				case headerNeed > len(m.inBuffer):
					m.header = append(m.header, m.inBuffer...)
					m.inBuffer = m.inBuffer[0:0]
					return false, nil
				case headerNeed <= len(m.inBuffer):
					m.header = append(m.header, m.inBuffer[:headerNeed]...)
					m.inBuffer = m.inBuffer[headerNeed:]
				}
				// Process header size
				if x, n := binary.Varint(m.header[:packetHeaderSize]); n <= 0 {
					log.Println("connReader header size error")
					return false, ErrPacketInvalid
				} else {
					m.size = int(x)
				}
				if m.size > packetSizeLimit {
					log.Println("connReader header oversize")
					return false, ErrPacketInvalid
				}
				// Body
				if m.size > 0 {
					m.body = make([]byte, 0, m.size)
				} else {
					m.header = m.header[0:0]
					m.body = []byte{}
					// new packet
					return true, nil
				}
			}
			if bodyNeed := m.size - len(m.body); bodyNeed > 0 {
				switch {
				case len(m.inBuffer) == 0:
					return false, nil
				case bodyNeed > len(m.inBuffer):
					m.body = append(m.body, m.inBuffer...)
					m.inBuffer = m.inBuffer[0:0]
					return false, nil
				case bodyNeed <= len(m.inBuffer):
					m.header = m.header[0:0]
					m.body = append(m.body, m.inBuffer[:bodyNeed]...)
					m.inBuffer = m.inBuffer[bodyNeed:]
					// new packet
					return true, nil
				}
			}
			return false, nil
		}()
		if err != nil {
			return buffers, err
		}
		if has {
			buffers = append(buffers, m.body)
		}
	}
	return buffers, err
}

//
// conn writer
//

type connWriter struct {
	conn   *connSafe
	header []byte
}

func (m *connWriter) init(conn *connSafe) {
	m.conn = conn
	m.header = make([]byte, packetHeaderSize)
}

func (m *connWriter) send(msg *ConnMessage) error {
	buf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	binary.PutVarint(m.header[:packetHeaderSize], int64(len(buf)))
	_, err = m.conn.Write(append(m.header, buf...))
	return err
}

//
// conn safe wrapper
//

type connSafe struct {
	net.Conn
	sync.Mutex
	closed bool
}

func (m *connSafe) safeClose() {
	m.Lock()
	if !m.closed {
		m.closed = true
		m.Close()
	}
	m.Unlock()
}

//
// conn message handler
//

func (m *ConnMessage) getType() ControlType {
	if m == nil || m.Control == nil {
		return ControlType_CUnknown
	}
	return m.Control.Type
}

func (m *ConnMessage) getAuthReq() *Auth_Request {
	if m.getType() != ControlType_CAuth {
		return nil
	}
	if m.Control.Auth == nil || m.Control.Auth.Req == nil {
		return nil
	}
	return m.Control.Auth.Req
}

func (m *ConnMessage) getAuthResp() *Auth_Response {
	if m.getType() != ControlType_CAuth {
		return nil
	}
	if m.Control.Auth == nil || m.Control.Auth.Resp == nil {
		return nil
	}
	return m.Control.Auth.Resp
}

func (m *ConnMessage) getSendNameReq() *SendName_Request {
	if m.getType() != ControlType_CSendName {
		return nil
	}
	if m.Control.SendName == nil || m.Control.SendName.Req == nil {
		return nil
	}
	return m.Control.SendName.Req
}

func (m *ConnMessage) getSendNameResp() *SendName_Response {
	if m.getType() != ControlType_CSendName {
		return nil
	}
	if m.Control.SendName == nil || m.Control.SendName.Resp == nil {
		return nil
	}
	return m.Control.SendName.Resp
}

func (m *ConnMessage) getAskNameReq() *AskName_Request {
	if m.getType() != ControlType_CAskName {
		return nil
	}
	if m.Control.AskName == nil || m.Control.AskName.Req == nil {
		return nil
	}
	return m.Control.AskName.Req
}

func (m *ConnMessage) getAskNameResp() *AskName_Response {
	if m.getType() != ControlType_CAskName {
		return nil
	}
	if m.Control.AskName == nil || m.Control.AskName.Resp == nil {
		return nil
	}
	return m.Control.AskName.Resp
}

func (m *ConnMessage) getGetNameReq() *GetName_Request {
	if m.getType() != ControlType_CGetName {
		return nil
	}
	if m.Control.GetName == nil || m.Control.GetName.Req == nil {
		return nil
	}
	return m.Control.GetName.Req
}

func (m *ConnMessage) getGetNameResp() *GetName_Response {
	if m.getType() != ControlType_CGetName {
		return nil
	}
	if m.Control.GetName == nil || m.Control.GetName.Resp == nil {
		return nil
	}
	return m.Control.GetName.Resp
}
