package go_actor

var Global global

type global struct {
}

func (m *global) Register(name string, actor *Actor) (id Id, err error) {
	return Id{}, nil
}
