package go_actor

var Global global

type global struct {
	node uint32
}

func (m *global) SelfNodeId() uint32 {
	return m.node
}

func (m *global) Register(name string, actor *Actor) (id Id, err error) {
	return Id{}, nil
}

//func (m *global) ByName(nameWrapper string) *GlobalRef {
//
//}
//
//func (m *global) Connect(addr string) error {
//
//}


type GlobalRef struct {
}

//func (m *GlobalRef) Send(sender *Id, msgContent ...interface{}) (err error) {
//
//}
//
//func (m *GlobalRef) Release() {
//
//}
