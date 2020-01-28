package go_actor

import (
	"sync"
	"time"
)

// session manager

type sessionManager struct {
	sync.Mutex
	currentId uint64
	sessions  map[uint64]*sessionWrapper
}

func (m *sessionManager) init() {
	m.sessions = map[uint64]*sessionWrapper{}
}

func (m *sessionManager) genSession() *sessionWrapper {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	for {
		m.currentId++
		if m.currentId == 0 {
			m.currentId++
		}
		if _, has := m.sessions[m.currentId]; !has {
			break
		}
	}
	w := &sessionWrapper{}
	w.init(m.currentId)
	m.sessions[m.currentId] = w
	return w
}

func (m *sessionManager) getSession(id uint64) *sessionWrapper {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	w, has := m.sessions[id]
	if !has {
		return nil
	}
	delete(m.sessions, id)
	return w
}

// session wrapper

type sessionWrapper struct {
	id        uint64
	msgCh     chan message
	createdAt time.Time
}

func (m *sessionWrapper) init(id uint64) {
	m.id = id
	m.msgCh = make(chan message, 1)
	m.createdAt = time.Now()
}
