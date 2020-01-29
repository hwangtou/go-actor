package go_actor

type globalManager struct {
	nodeId uint32
}

func (m *globalManager) init() {
}

func (m *globalManager) Register(name string, actor *Actor) (id Id, err error) {
	return Id{}, nil
}

//func (m *globalManager) ByName(nameWrapper string) *GlobalRef {
//
//}
//
//func (m *globalManager) Connect(addr string) error {
//
//}

//
// Remote Ref
//

type RemoteRef struct {
	id Id
}

func (m RemoteRef) Id() Id {
	return m.id
}

func (m *RemoteRef) Send(sender Ref, msg interface{}) error {
	panic("implement me")
}

func (m *RemoteRef) Ask(sender Ref, ask interface{}) (answer interface{}, err error) {
	panic("implement me")
}

func (m *RemoteRef) Shutdown(sender Ref) error {
	panic("implement me")
}

