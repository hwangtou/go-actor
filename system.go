package actor

import "regexp"

//
// PRIVATE
//

func init() {
	defaultSys = NewSystem()
	Remote = &defaultSys.remote
}

// Go-actor provides a default system instance for use.
// Developer can create system instance if needed, but not recommended.
var defaultSys *system

// It's the core of go-actor.
type system struct {
	locals localsManager
	remote remoteManager
}

// Developer can create system instance if needed, but not recommended.
func NewSystem() *system {
	defaultSys = &system{}
	defaultSys.init()
	return defaultSys
}

func (m *system) init() {
	m.locals.init(m)
	m.remote.init(m)
}

func (m *system) Spawn(fn func() Actor, arg interface{}) (*LocalRef, error) {
	return m.locals.spawnActor(fn, "", arg)
}

func (m *system) SpawnWithName(fn func() Actor, name string, arg interface{}) (*LocalRef, error) {
	return m.locals.spawnActor(fn, name, arg)
}

func (m *system) Register(ref Ref, name string) error {
	lr, ok := ref.(*LocalRef)
	if !ok {
		return ErrNotLocalActor
	}
	return m.locals.setNameRunning(lr, name)
}

func (m *system) ById(id uint32) *LocalRef {
	return m.locals.getActorRef(id)
}

func (m *system) ByName(name string) *LocalRef {
	return m.locals.getName(name)
}

func (m *system) SearchName(name string) map[string]*LocalRef {
	result := map[string]*LocalRef{}
	reg := regexp.MustCompile(regexp.QuoteMeta(name))
	for actorName, wrapper := range m.locals.names {
		if reg.FindString(actorName) != "" {
			localRef := m.locals.getActorRef(wrapper.id)
			if localRef != nil {
				result[actorName] = localRef
			}
		}
	}
	return result
}

func (m *system) Count() int {
	return len(m.locals.actors)
}

func (m *system) Remote() *remoteManager {
	return &m.remote
}
