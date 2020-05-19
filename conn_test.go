package actor

import (
	"net"
	"reflect"
	"sync"
	"testing"
)

func Test_connReader_handleBuffer(t *testing.T) {
	type fields struct {
		conn     *connSafe
		recvCh   chan *ConnMessage
		inBuffer []byte
		header   []byte
		body     []byte
		size     int
	}
	type args struct {
		buf []byte
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantBuffers [][]byte
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connReader{
				conn:     tt.fields.conn,
				recvCh:   tt.fields.recvCh,
				inBuffer: tt.fields.inBuffer,
				header:   tt.fields.header,
				body:     tt.fields.body,
				size:     tt.fields.size,
			}
			gotBuffers, err := m.handleBuffer(tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleBuffer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBuffers, tt.wantBuffers) {
				t.Errorf("handleBuffer() gotBuffers = %v, want %v", gotBuffers, tt.wantBuffers)
			}
		})
	}
}

func Test_connReader_loop(t *testing.T) {
	type fields struct {
		conn     *connSafe
		recvCh   chan *ConnMessage
		inBuffer []byte
		header   []byte
		body     []byte
		size     int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connReader{
				conn:     tt.fields.conn,
				recvCh:   tt.fields.recvCh,
				inBuffer: tt.fields.inBuffer,
				header:   tt.fields.header,
				body:     tt.fields.body,
				size:     tt.fields.size,
			}
		})
	}
}

func Test_connSafe_safeClose(t *testing.T) {
	type fields struct {
		Conn   net.Conn
		Mutex  sync.Mutex
		closed bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connSafe{
				Conn:   tt.fields.Conn,
				Mutex:  tt.fields.Mutex,
				closed: tt.fields.closed,
			}
		})
	}
}

func Test_connWriter_send(t *testing.T) {
	type fields struct {
		conn   *connSafe
		header []byte
	}
	type args struct {
		msg *ConnMessage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connWriter{
				conn:   tt.fields.conn,
				header: tt.fields.header,
			}
			if err := m.send(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_conn_close(t *testing.T) {
	type fields struct {
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
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &conn{
				remote:      tt.fields.remote,
				ready:       tt.fields.ready,
				listener:    tt.fields.listener,
				inAuth:      tt.fields.inAuth,
				inConn:      tt.fields.inConn,
				inConnLock:  tt.fields.inConnLock,
				inMessageCh: tt.fields.inMessageCh,
				outConn:     tt.fields.outConn,
				outConnLock: tt.fields.outConnLock,
			}
		})
	}
}

func Test_conn_getOutConn(t *testing.T) {
	type fields struct {
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
	type args struct {
		nodeId uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *outNode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &conn{
				remote:      tt.fields.remote,
				ready:       tt.fields.ready,
				listener:    tt.fields.listener,
				inAuth:      tt.fields.inAuth,
				inConn:      tt.fields.inConn,
				inConnLock:  tt.fields.inConnLock,
				inMessageCh: tt.fields.inMessageCh,
				outConn:     tt.fields.outConn,
				outConnLock: tt.fields.outConnLock,
			}
			if got := m.getOutConn(tt.args.nodeId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOutConn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_conn_getOutConnOrDial(t *testing.T) {
	type fields struct {
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
	type args struct {
		nodeId uint32
		auth   string
		nw     Network
		addr   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *outNode
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &conn{
				remote:      tt.fields.remote,
				ready:       tt.fields.ready,
				listener:    tt.fields.listener,
				inAuth:      tt.fields.inAuth,
				inConn:      tt.fields.inConn,
				inConnLock:  tt.fields.inConnLock,
				inMessageCh: tt.fields.inMessageCh,
				outConn:     tt.fields.outConn,
				outConnLock: tt.fields.outConnLock,
			}
			got, err := m.getOutConnOrDial(tt.args.nodeId, tt.args.auth, tt.args.nw, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOutConnOrDial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOutConnOrDial() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_conn_inConnHandler(t *testing.T) {
	type fields struct {
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
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &conn{
				remote:      tt.fields.remote,
				ready:       tt.fields.ready,
				listener:    tt.fields.listener,
				inAuth:      tt.fields.inAuth,
				inConn:      tt.fields.inConn,
				inConnLock:  tt.fields.inConnLock,
				inMessageCh: tt.fields.inMessageCh,
				outConn:     tt.fields.outConn,
				outConnLock: tt.fields.outConnLock,
			}
		})
	}
}

func Test_conn_inMessageHandler(t *testing.T) {
	type fields struct {
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
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &conn{
				remote:      tt.fields.remote,
				ready:       tt.fields.ready,
				listener:    tt.fields.listener,
				inAuth:      tt.fields.inAuth,
				inConn:      tt.fields.inConn,
				inConnLock:  tt.fields.inConnLock,
				inMessageCh: tt.fields.inMessageCh,
				outConn:     tt.fields.outConn,
				outConnLock: tt.fields.outConnLock,
			}
		})
	}
}

func Test_inReply_reply(t *testing.T) {
	type fields struct {
		inMessage *ConnMessage
		inConn    *inNode
	}
	type args struct {
		message *ConnMessage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &inReply{
				inMessage: tt.fields.inMessage,
				inConn:    tt.fields.inConn,
			}
			if err := m.reply(tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("reply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_outNode_auth(t *testing.T) {
	type fields struct {
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
	type args struct {
		nodeId   uint32
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &outNode{
				ready:   tt.fields.ready,
				global:  tt.fields.global,
				nodeId:  tt.fields.nodeId,
				nw:      tt.fields.nw,
				addr:    tt.fields.addr,
				conn:    tt.fields.conn,
				reader:  tt.fields.reader,
				writer:  tt.fields.writer,
				seq:     tt.fields.seq,
				seqId:   tt.fields.seqId,
				seqLock: tt.fields.seqLock,
			}
			if err := m.auth(tt.args.nodeId, tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("auth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_outNode_close(t *testing.T) {
	type fields struct {
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
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &outNode{
				ready:   tt.fields.ready,
				global:  tt.fields.global,
				nodeId:  tt.fields.nodeId,
				nw:      tt.fields.nw,
				addr:    tt.fields.addr,
				conn:    tt.fields.conn,
				reader:  tt.fields.reader,
				writer:  tt.fields.writer,
				seq:     tt.fields.seq,
				seqId:   tt.fields.seqId,
				seqLock: tt.fields.seqLock,
			}
		})
	}
}

func Test_outNode_dial(t *testing.T) {
	type fields struct {
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
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &outNode{
				ready:   tt.fields.ready,
				global:  tt.fields.global,
				nodeId:  tt.fields.nodeId,
				nw:      tt.fields.nw,
				addr:    tt.fields.addr,
				conn:    tt.fields.conn,
				reader:  tt.fields.reader,
				writer:  tt.fields.writer,
				seq:     tt.fields.seq,
				seqId:   tt.fields.seqId,
				seqLock: tt.fields.seqLock,
			}
			if err := m.dial(tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("dial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_outNode_loop(t *testing.T) {
	type fields struct {
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
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &outNode{
				ready:   tt.fields.ready,
				global:  tt.fields.global,
				nodeId:  tt.fields.nodeId,
				nw:      tt.fields.nw,
				addr:    tt.fields.addr,
				conn:    tt.fields.conn,
				reader:  tt.fields.reader,
				writer:  tt.fields.writer,
				seq:     tt.fields.seq,
				seqId:   tt.fields.seqId,
				seqLock: tt.fields.seqLock,
			}
		})
	}
}

func Test_outNode_send(t *testing.T) {
	type fields struct {
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
	type args struct {
		message *ConnMessage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *seqWrapper
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &outNode{
				ready:   tt.fields.ready,
				global:  tt.fields.global,
				nodeId:  tt.fields.nodeId,
				nw:      tt.fields.nw,
				addr:    tt.fields.addr,
				conn:    tt.fields.conn,
				reader:  tt.fields.reader,
				writer:  tt.fields.writer,
				seq:     tt.fields.seq,
				seqId:   tt.fields.seqId,
				seqLock: tt.fields.seqLock,
			}
			got, err := m.send(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("send() got = %v, want %v", got, tt.want)
			}
		})
	}
}