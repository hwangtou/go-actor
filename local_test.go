package actor

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestLocalRef_Ask(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
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
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
			if err := m.Ask(tt.args.sender, tt.args.ask, tt.args.answer); (err != nil) != tt.wantErr {
				t.Errorf("Ask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalRef_Id(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
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
			m := LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
			if got := m.Id(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalRef_Send(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
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
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
			if err := m.Send(tt.args.sender, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalRef_Shutdown(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
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
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
			if err := m.Shutdown(tt.args.sender); (err != nil) != tt.wantErr {
				t.Errorf("Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalRef_checkStatus(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
	}
	type args struct {
		status Status
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
			if got := m.checkStatus(tt.args.status); got != tt.want {
				t.Errorf("checkStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalRef_logMessageError(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
	}
	type args struct {
		err error
		msg message
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
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
		})
	}
}

func TestLocalRef_receiving(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
	}
	type args struct {
		msg *message
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
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
			if err := m.receiving(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("receiving() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalRef_setStatus(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
	}
	type args struct {
		status Status
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
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
		})
	}
}

func TestLocalRef_spawn(t *testing.T) {
	type fields struct {
		local       *localsManager
		id          Id
		status      Status
		statusLock  sync.RWMutex
		actor       Actor
		ask         Ask
		recvCh      chan *message
		recvRunning bool
		recvBeginAt time.Time
		recvEndAt   time.Time
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LocalRef{
				local:       tt.fields.local,
				id:          tt.fields.id,
				status:      tt.fields.status,
				statusLock:  tt.fields.statusLock,
				actor:       tt.fields.actor,
				ask:         tt.fields.ask,
				recvCh:      tt.fields.recvCh,
				recvRunning: tt.fields.recvRunning,
				recvBeginAt: tt.fields.recvBeginAt,
				recvEndAt:   tt.fields.recvEndAt,
			}
		})
	}
}

func Test_localsManager_delActorRef(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		id uint32
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
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
		})
	}
}

func Test_localsManager_getActorRef(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		id uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *LocalRef
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
			if got := m.getActorRef(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getActorRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localsManager_getName(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *LocalRef
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
			if got := m.getName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localsManager_newActorRef(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		a Actor
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *LocalRef
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
			if got := m.newActorRef(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newActorRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localsManager_setNameRunning(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		ref  *LocalRef
		name string
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
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
			if err := m.setNameRunning(tt.args.ref, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("setNameRunning() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_localsManager_setNameSpawn(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		ref    *LocalRef
		name   string
		status Status
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
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
			if err := m.setNameSpawn(tt.args.ref, tt.args.name, tt.args.status); (err != nil) != tt.wantErr {
				t.Errorf("setNameSpawn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_localsManager_shutdownActor(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		r *LocalRef
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
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
		})
	}
}

func Test_localsManager_spawnActor(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		fn   func() Actor
		name string
		arg  interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *LocalRef
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
			got, err := m.spawnActor(tt.args.fn, tt.args.name, tt.args.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("spawnActor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("spawnActor() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localsManager_unsetNameSpawn(t *testing.T) {
	type fields struct {
		sys         *system
		sessions    sessionsManager
		actors      map[uint32]*LocalRef
		idCount     uint32
		idCountLock sync.Mutex
		names       map[string]nameWrapper
		namesLock   sync.RWMutex
	}
	type args struct {
		ref    *LocalRef
		name   string
		status Status
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
			m := &localsManager{
				sys:         tt.fields.sys,
				sessions:    tt.fields.sessions,
				actors:      tt.fields.actors,
				idCount:     tt.fields.idCount,
				idCountLock: tt.fields.idCountLock,
				names:       tt.fields.names,
				namesLock:   tt.fields.namesLock,
			}
		})
	}
}

func Test_sessionsManager_handleSession(t *testing.T) {
	type fields struct {
		Mutex     sync.Mutex
		currentId uint64
		sessions  map[uint64]*session
	}
	type args struct {
		id     uint64
		answer message
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
			m := &sessionsManager{
				Mutex:     tt.fields.Mutex,
				currentId: tt.fields.currentId,
				sessions:  tt.fields.sessions,
			}
		})
	}
}

func Test_sessionsManager_newSession(t *testing.T) {
	type fields struct {
		Mutex     sync.Mutex
		currentId uint64
		sessions  map[uint64]*session
	}
	tests := []struct {
		name   string
		fields fields
		want   *session
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &sessionsManager{
				Mutex:     tt.fields.Mutex,
				currentId: tt.fields.currentId,
				sessions:  tt.fields.sessions,
			}
			if got := m.newSession(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSession() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sessionsManager_popSession(t *testing.T) {
	type fields struct {
		Mutex     sync.Mutex
		currentId uint64
		sessions  map[uint64]*session
	}
	type args struct {
		id uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *session
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &sessionsManager{
				Mutex:     tt.fields.Mutex,
				currentId: tt.fields.currentId,
				sessions:  tt.fields.sessions,
			}
			if got := m.popSession(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("popSession() = %v, want %v", got, tt.want)
			}
		})
	}
}