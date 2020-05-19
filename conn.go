// Copyright 2020 Tou.Hwang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package actor

import (
	"encoding/binary"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
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

// Connection
type conn struct {
	remote      *remoteManager
	ready       bool
	listener    net.Listener
	inAuth      string
	inConn      map[uint32]*inNode
	inConnLock  sync.RWMutex
	inMessageCh chan *inReply
	outConn     map[uint32]*outNode
	outConnLock sync.RWMutex
}

func (m *conn) init(nw Network, listen string) error {
	l, err := net.Listen(string(nw), listen)
	if err != nil {
		m.ready = false
		log.Println("actor.Remote.init error,", err)
		return err
	}
	m.ready = true
	m.listener = l
	m.inAuth = ""
	m.inConn = make(map[uint32]*inNode)
	m.inMessageCh = make(chan *inReply, inMessageChannelLength)
	m.outConn = make(map[uint32]*outNode)
	go m.inMessageHandler()
	go m.inConnHandler()
	return nil
}

func (m *conn) inMessageHandler() {
	log.Println("actor.Remote starts incoming message handle loop")
	for {
		msg, more := <-m.inMessageCh
		if !more {
			break
		}
		if msg == nil || msg.inConn == nil || msg.inMessage == nil ||
			msg.inMessage.Content == nil {
			log.Println("actor.Remote handled incoming message error,", msg)
			continue
		}
		// High effective parallel working mode
		go func() {
			replyMessage := &ConnMessage{}
			switch msg.inMessage.Type {
			case ControlType_CSendName:
				{
					// Validation
					sendNameWrapper := msg.inMessage.GetSendName()
					resp := &SendName_Response{
						HasError: true,
					}
					replyMessage.Type = ControlType_CSendName
					replyMessage.Content = &ConnMessage_SendName{
						SendName: &SendName{
							Data: &SendName_Resp{
								Resp: resp,
							},
						},
					}
					if sendNameWrapper == nil {
						log.Println("actor.Remote handled incoming message, empty send message error,", msg)
						resp.ErrorMessage = "Empty message"
						break
					}
					sendName := sendNameWrapper.GetReq()
					if sendName == nil || sendName.SendData == nil {
						log.Println("actor.Remote handled incoming message, empty send message request error,", msg)
						resp.ErrorMessage = "Empty message request"
						break
					}

					// Get local actor by name
					localRef := m.remote.sys.locals.getName(sendName.ToName)
					if localRef == nil {
						resp.ErrorMessage = "Actor name not found"
						break
					}

					// Process send message
					var (
						sendFromRef *RemoteRef
						sendMessage interface{}
						sendError   error
					)
					if sendName.FromId != 0 {
						sendFromRef = &RemoteRef{
							id: Id{
								node: msg.inConn.nodeId,
								id:   sendName.FromId,
								name: sendName.FromName,
							},
							node: nil, // todo
						}
					}
					switch sendName.SendData.Type {
					case DataType_ProtoBuf:
						sendData := sendName.SendData.GetProto()
						sendDataProto, err := ptypes.Empty(sendData)
						if err != nil {
							sendError = err
							break
						}
						err = ptypes.UnmarshalAny(sendData, sendDataProto)
						if err != nil {
							sendError = err
							break
						}
						sendMessage = sendDataProto
					case DataType_String:
						sendMessage = sendName.SendData.GetStr()
					default:
						sendError = errors.New("unsupported type") // todo
						break
					}
					if sendError != nil {
						resp.ErrorMessage = sendError.Error()
						break
					}

					// Send local actor
					sendError = localRef.Send(sendFromRef, sendMessage)
					if sendError != nil {
						resp.ErrorMessage = sendError.Error()
					} else {
						resp.HasError = false
					}
				}
			case ControlType_CAskName:
				{
					// Validation
					askNameWrapper := msg.inMessage.GetAskName()
					resp := &AskName_Response{
						HasError: true,
						AnswerData: &DataContentType{},
					}
					replyMessage.Type = ControlType_CAskName
					replyMessage.Content = &ConnMessage_AskName{
						AskName: &AskName{
							Data: &AskName_Resp{
								Resp: resp,
							},
						},
					}
					if askNameWrapper == nil {
						log.Println("actor.Remote handled incoming message, empty ask message error,", msg)
						resp.ErrorMessage = "Empty message"
						break
					}
					askName := askNameWrapper.GetReq()
					if askName == nil || askName.AskData == nil || askName.AnswerData == nil {
						log.Println("actor.Remote handled incoming message, empty ask message request error,", msg)
						resp.ErrorMessage = "Empty message request"
						break
					}

					// Get local actor by name
					localRef := m.remote.sys.locals.getName(askName.ToName)
					if localRef == nil {
						resp.ErrorMessage = "Actor name not found"
						break
					}

					// Process ask message
					var (
						askFromRef    *RemoteRef
						askMessage    interface{}
						answerError   error
					)
					if askName.FromId != 0 {
						askFromRef = &RemoteRef{
							id: Id{
								node: msg.inConn.nodeId,
								id:   askName.FromId,
								name: askName.FromName,
							},
							node: nil, // todo
						}
					}
					// Ask
					switch askName.AskData.Type {
					case DataType_ProtoBuf:
						askData := askName.AskData.GetProto()
						askDataProto, err := ptypes.Empty(askData)
						if err != nil {
							answerError = err
							break
						}
						err = ptypes.UnmarshalAny(askData, askDataProto)
						if err != nil {
							answerError = err
							break
						}
						askMessage = askDataProto
					case DataType_String:
						askMessage = askName.AskData.GetStr()
					default:
						answerError = errors.New("unsupported type") // todo
						break
					}
					if answerError != nil {
						resp.ErrorMessage = answerError.Error()
						break
					}
					// Answer
					switch askName.AnswerData.Type {
					case DataType_ProtoBuf:
						var answerProto *any.Any
						resp.AnswerData.Type = DataType_ProtoBuf
						emptyAnswerProto := askName.AnswerData.GetProto()
						answerInstance, err := ptypes.Empty(emptyAnswerProto)
						if err != nil {
							answerError = err
							break
						}
						err = ptypes.UnmarshalAny(emptyAnswerProto, answerInstance)
						if err != nil {
							answerError = err
							break
						}
						// Send local actor
						answerError = localRef.Ask(askFromRef, askMessage, &answerInstance)
						answerProto, err = ptypes.MarshalAny(answerInstance)
						if err != nil {
							answerError = err
							break
						}
						resp.AnswerData.Content = &DataContentType_Proto{
							Proto: answerProto,
						}
					case DataType_String:
						answerString := ""
						resp.AnswerData.Type = DataType_String
						// Send local actor
						answerError = localRef.Ask(askFromRef, askMessage, &answerString)
						resp.AnswerData.Content = &DataContentType_Str{
							Str: answerString,
						}
					default:
						answerError = errors.New("unsupported type") // todo
						break
					}
					// Error
					if answerError != nil {
						resp.ErrorMessage = answerError.Error()
					} else {
						resp.HasError = false
					}
				}
			case ControlType_CGetName:
				{
					// Validation
					getNameWrapper := msg.inMessage.GetGetName()
					resp := &GetName_Response{
						Has: false,
					}
					replyMessage.Type = ControlType_CGetName
					replyMessage.Content = &ConnMessage_GetName{
						GetName: &GetName{
							Data: &GetName_Resp{
								Resp: resp,
							},
						},
					}
					if getNameWrapper == nil {
						log.Println("actor.Remote handled incoming message, empty get message error,", msg)
						resp.ErrorMessage = "Empty message"
						break
					}
					getName := getNameWrapper.GetReq()
					if getName == nil || getName.Name == "" {
						log.Println("actor.Remote handled incoming message, empty get message request error,", msg)
						resp.ErrorMessage = "Empty message request"
						break
					}

					// Process get message
					if lr := m.remote.sys.locals.getName(getName.Name); lr != nil {
						resp.Has = true
						resp.ActorId = lr.id.id
					} else {
						resp.ErrorMessage = "Actor name not found"
					}
				}
			default:
				log.Println("actor.Remote handled incoming message type error,", msg)
			}
			if err := msg.reply(replyMessage); err != nil {
				log.Println("actor.Remote handled incoming message reply error,", err)
			}
		}()
	}
}

func (m *conn) inConnHandler() {
	log.Println("actor.Remote starts incoming connection handle loop")
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			log.Println("actor.Remote incoming connection accepted failed,", err)
			break
		}
		go func() {
			n := &inNode{
				ready:  true,
				global: m.remote,
				nodeId: 0,
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
			isAuth, sequenceId := false, uint64(0)
			select {
			case packet, more := <-n.reader.recvCh:
				{
					if !more {
						log.Println("actor.Remote.listen auth no message", packet) // todo: desc
						// todo reply error
						break
					}
					if packet == nil || packet.GetAuth() == nil || packet.GetAuth().GetReq() == nil {
						log.Println("actor.Remote.listen auth info error", packet) // todo: desc
						// todo reply error
						break
					}
					req := packet.GetAuth().GetReq()
					if req.Password != m.inAuth {
						log.Println("actor.Remote.listen auth invalid,", packet) // todo: desc
						// todo reply error
						break
					}
					if req.FromNodeId == 0 {
						log.Println("actor.Remote.listen auth invalid,", packet) // todo: desc
						// todo reply error
						break
					}
					if req.ToNodeId != m.remote.nodeId {
						log.Println("actor.Remote.listen auth invalid,", packet) // todo: desc
						// todo reply error
						break
					}
					n.nodeId = req.FromNodeId
					if oldN, has := m.inConn[n.nodeId]; has {
						oldN.conn.safeClose()
					}
					m.inConn[n.nodeId] = n
					sequenceId = packet.SequenceId
					// todo reply okay
					isAuth = true
					// todo binary dial
				}
			case <-time.After(authTimeout * time.Second):
				log.Println("actor.Remote.listen auth timeout") // todo: desc
				return
			}
			if err := n.writer.send(&ConnMessage{
				SequenceId: sequenceId,
				Type:       ControlType_CAuth,
				Direction:  Direction_Response,
				Content: &ConnMessage_Auth{
					Auth: &Auth{
						Data: &Auth_Resp{
							Resp: &Auth_Response{
								IsAuth: isAuth,
							},
						},
					},
				},
			}); err != nil {
				log.Println("actor.Remote.listen send error,") // todo: desc
			}
			if !isAuth {
				return
			}
			// reply
			for {
				packet, more := <-n.reader.recvCh
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

func (m *conn) getOutConn(nodeId uint32) *outNode {
	if nodeId == 0 {
		return nil
	}

	// get conn
	m.outConnLock.RLock()
	n, _ := m.outConn[nodeId]
	m.outConnLock.RUnlock()
	return n
}

func (m *conn) getOutConnOrDial(nodeId uint32, auth string, nw Network, addr string) (*outNode, error) {
	if nodeId == 0 {
		return nil, ErrNodeId
	}

	m.outConnLock.Lock()
	defer m.outConnLock.Unlock()

	// get conn
	n, has := m.outConn[nodeId]
	if has {
		return n, nil
	}

	// todo
	n = &outNode{
		ready:   false,
		global:  m.remote,
		nodeId:  nodeId,
		nw:      nw,
		addr:    addr,
		conn:    connSafe{},
		reader:  connReader{},
		writer:  connWriter{},
		seq:     make(map[uint64]*seqWrapper),
		seqId:   0,
		seqLock: sync.Mutex{},
	}
	m.outConn[nodeId] = n
	if err := n.dial(auth); err != nil {
		delete(m.outConn, nodeId)
		return nil, err
	}
	log.Println("actor.Remote.getOutConnOrDial new conn")
	return n, nil
}

func (m *conn) close() {
	m.ready = false
	m.listener.Close()
	m.inAuth = ""
	m.inConnLock.Lock()
	for nodeId, node := range m.inConn {
		node.conn.safeClose()
		delete(m.inConn, nodeId)
	}
	m.inConnLock.Unlock()
	m.outConnLock.Lock()
	for nodeId, node := range m.outConn {
		node.conn.safeClose()
		for id, seq := range node.seq {
			seq.respCh <- nil
			delete(node.seq, id)
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
	global  *remoteManager
	nodeId  uint32
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

func (m *outNode) dial(password string) (err error) {
	m.ready = false
	m.conn.closed = false
	m.conn.Conn, err = net.Dial(string(m.nw), m.addr)
	if err != nil {
		log.Println("actor.Remote.getOutConnOrDial dial error,", err)
		return err
	}
	m.reader.init(&m.conn)
	m.writer.init(&m.conn)

	err = m.auth(m.nodeId, password)
	if err != nil {
		log.Println("actor.Remote.getOutConnOrDial password error,", err)
		m.conn.safeClose()
		return err
	}

	m.ready = true
	go m.loop()
	return nil
}

func (m *outNode) auth(nodeId uint32, password string) error {
	// send authentication
	if err := m.writer.send(&ConnMessage{
		SequenceId: 0,
		Type:       ControlType_CAuth,
		Direction:  Direction_Request,
		Content: &ConnMessage_Auth{
			Auth: &Auth{
				Data: &Auth_Req{
					Req: &Auth_Request{
						FromNodeId: m.global.nodeId,
						ToNodeId:   nodeId,
						Password:   password,
					},
				},
			},
		},
	}); err != nil {
		return err
	}
	// receive
	select {
	case packet, more := <-m.reader.recvCh:
		if !more {
			return ErrConnError
		}
		if packet == nil || packet.GetAuth() == nil || packet.GetAuth().GetResp() == nil {
			return ErrAuthFailed
		}
		resp := packet.GetAuth().GetResp()
		if resp.IsAuth {
			return nil
		}
		log.Println("actor.Remote.getOutConnOrDial auth error message", packet) // todo desc
		return ErrAuthFailed
	case <-time.After(authTimeout * time.Second):
		return ErrAuthTimeout
	}
}

func (m *outNode) loop() {
	for {
		packet, more := <-m.reader.recvCh
		if !more {
			return
		}
		seq, has := m.seq[packet.SequenceId]
		if !has {
			log.Println("actor.Remote out node receive unknown sequence,", packet)
			continue
		}
		m.seqLock.Lock()
		delete(m.seq, packet.SequenceId)
		m.seqLock.Unlock()
		seq.respCh <- packet
	}
}

func (m *outNode) send(message *ConnMessage) (*seqWrapper, error) {
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
	message.SequenceId = seqId
	message.Direction = Direction_Request
	w := &seqWrapper{
		req:       message,
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
	global *remoteManager
	nodeId uint32
	conn   connSafe
	reader connReader
	writer connWriter
}

type inReply struct {
	inMessage *ConnMessage
	inConn    *inNode
}

func (m *inReply) reply(message *ConnMessage) error {
	if !m.inConn.ready {
		return ErrReplyFailed
	}
	message.SequenceId = m.inMessage.SequenceId
	message.Direction = Direction_Response
	return m.inConn.writer.send(message)
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
					log.Println("actor.Remote out conn loop error,", err)
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

//func (m *ConnMessage) getType() ControlType {
//	if m == nil || m.Control == nil {
//		return ControlType_CUnknown
//	}
//	return m.Control.Type
//}
//
//func (m *ConnMessage) getAuthReq() *Auth_Request {
//	if m.getType() != ControlType_CAuth {
//		return nil
//	}
//	if m.Control.Auth == nil || m.Control.Auth.Req == nil {
//		return nil
//	}
//	return m.Control.Auth.Req
//}
//
//func (m *ConnMessage) getAuthResp() *Auth_Response {
//	if m.getType() != ControlType_CAuth {
//		return nil
//	}
//	if m.Control.Auth == nil || m.Control.Auth.Resp == nil {
//		return nil
//	}
//	return m.Control.Auth.Resp
//}
//
//func (m *ConnMessage) getSendNameReq() *SendName_Request {
//	if m.getType() != ControlType_CSendName {
//		return nil
//	}
//	if m.Control.SendName == nil || m.Control.SendName.Req == nil {
//		return nil
//	}
//	return m.Control.SendName.Req
//}
//
//func (m *ConnMessage) getSendNameResp() *SendName_Response {
//	if m.getType() != ControlType_CSendName {
//		return nil
//	}
//	if m.Control.SendName == nil || m.Control.SendName.Resp == nil {
//		return nil
//	}
//	return m.Control.SendName.Resp
//}
//
//func (m *ConnMessage) getAskNameReq() *AskName_Request {
//	if m.getType() != ControlType_CAskName {
//		return nil
//	}
//	if m.Control.AskName == nil || m.Control.AskName.Req == nil {
//		return nil
//	}
//	return m.Control.AskName.Req
//}
//
//func (m *ConnMessage) getAskNameResp() *AskName_Response {
//	if m.getType() != ControlType_CAskName {
//		return nil
//	}
//	if m.Control.AskName == nil || m.Control.AskName.Resp == nil {
//		return nil
//	}
//	return m.Control.AskName.Resp
//}
//
//func (m *ConnMessage) getGetNameReq() *GetName_Request {
//	if m.getType() != ControlType_CGetName {
//		return nil
//	}
//	if m.Control.GetName == nil || m.Control.GetName.Req == nil {
//		return nil
//	}
//	return m.Control.GetName.Req
//}
//
//func (m *ConnMessage) getGetNameResp() *GetName_Response {
//	if m.getType() != ControlType_CGetName {
//		return nil
//	}
//	if m.Control.GetName == nil || m.Control.GetName.Resp == nil {
//		return nil
//	}
//	return m.Control.GetName.Resp
//}
