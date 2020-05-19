package actor

import (
	"reflect"
	"testing"
)

func TestRemoteConn_ByName(t *testing.T) {
	type fields struct {
		node *outNode
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *RemoteRef
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RemoteConn{
				node: tt.fields.node,
			}
			got, err := m.ByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteRef_Ask(t *testing.T) {
	type fields struct {
		id   Id
		node *outNode
	}
	type args struct {
		sender Ref
		ask    interface{}
		answer interface{}
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
			m := &RemoteRef{
				id:   tt.fields.id,
				node: tt.fields.node,
			}
			if err := m.Ask(tt.args.sender, tt.args.ask, tt.args.answer); (err != nil) != tt.wantErr {
				t.Errorf("Ask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteRef_Id(t *testing.T) {
	type fields struct {
		id   Id
		node *outNode
	}
	tests := []struct {
		name   string
		fields fields
		want   Id
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := RemoteRef{
				id:   tt.fields.id,
				node: tt.fields.node,
			}
			if got := m.Id(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoteRef_Send(t *testing.T) {
	type fields struct {
		id   Id
		node *outNode
	}
	type args struct {
		sender Ref
		msg    interface{}
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
			m := &RemoteRef{
				id:   tt.fields.id,
				node: tt.fields.node,
			}
			if err := m.Send(tt.args.sender, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteRef_Shutdown(t *testing.T) {
	type fields struct {
		id   Id
		node *outNode
	}
	type args struct {
		sender Ref
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
			m := &RemoteRef{
				id:   tt.fields.id,
				node: tt.fields.node,
			}
			if err := m.Shutdown(tt.args.sender); (err != nil) != tt.wantErr {
				t.Errorf("Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}