package actor

import (
	"log"
	"sync"
	"time"
)

// session manager

type sessionsManager struct {
	sync.Mutex
	currentId uint64
	sessions  map[uint64]*session
}

func (m *sessionsManager) init() {
	m.sessions = map[uint64]*session{}
}

func (m *sessionsManager) newSession() *session {
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
	w := &session{}
	w.init(m.currentId)
	m.sessions[m.currentId] = w
	return w
}

func (m *sessionsManager) popSession(id uint64) *session {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	w, has := m.sessions[id]
	if !has {
		return nil
	}
	delete(m.sessions, id)
	return w
}

func (m *sessionsManager) handleSession(id uint64, answer message) {
	s := m.popSession(id)
	if s == nil {
		log.Println("SERIOUS! Session not found!", id, answer)
		return
	}

	// TODO handle remote message
	defer func() {
		if recover() != nil {
			log.Println("oops handle closed session")
		}
	}()
	s.msgCh <- answer
}

// session wrapper

type session struct {
	id        uint64
	msgCh     chan message
	createdAt time.Time
}

func (m *session) init(id uint64) {
	m.id = id
	m.msgCh = make(chan message, 1)
	m.createdAt = time.Now()
}
