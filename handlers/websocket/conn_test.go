package websocket

import (
	"github.com/hwangtou/go-actor"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func Test_connection_HandleAsk(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	type args struct {
		sender actor.Ref
		ask    interface{}
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantAnswer interface{}
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
			gotAnswer, err := m.HandleAsk(tt.args.sender, tt.args.ask)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleAsk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAnswer, tt.wantAnswer) {
				t.Errorf("HandleAsk() gotAnswer = %v, want %v", gotAnswer, tt.wantAnswer)
			}
		})
	}
}

func Test_connection_HandleSend(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	type args struct {
		sender  actor.Ref
		message interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
		})
	}
}

func Test_connection_Shutdown(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
		})
	}
}

func Test_connection_StartUp(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	type args struct {
		self *actor.LocalRef
		arg  interface{}
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
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
			if err := m.StartUp(tt.args.self, tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("StartUp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connection_Started(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
		})
	}
}

func Test_connection_Type(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	tests := []struct {
		name        string
		fields      fields
		wantName    string
		wantVersion int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
			gotName, gotVersion := m.Type()
			if gotName != tt.wantName {
				t.Errorf("Type() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("Type() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

func Test_connection_changeForwardingActor(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	type args struct {
		forwarding actor.Ref
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
		})
	}
}

func Test_connection_sendMessage(t *testing.T) {
	type fields struct {
		self           *actor.LocalRef
		conn           *websocket.Conn
		header         http.Header
		readTimeout    time.Time
		writeTimeout   time.Time
		forwarding     actor.Ref
		acceptedOrDial bool
	}
	type args struct {
		t   int
		buf []byte
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
			m := &connection{
				self:           tt.fields.self,
				conn:           tt.fields.conn,
				header:         tt.fields.header,
				readTimeout:    tt.fields.readTimeout,
				writeTimeout:   tt.fields.writeTimeout,
				forwarding:     tt.fields.forwarding,
				acceptedOrDial: tt.fields.acceptedOrDial,
			}
			if err := m.sendMessage(tt.args.t, tt.args.buf); (err != nil) != tt.wantErr {
				t.Errorf("sendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
