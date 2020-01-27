package go_actor

import (
	"log"
	"sync"
	"time"
)

//
// message
//

type message struct {
	sender     Ref
	ask        Ref
	sequence   uint64
	msgType    int
	msgContent interface{}
	msgError   error
	isReply    bool
}

type sequence struct {
	sync.Mutex
	ask ActorAsk
	id uint64
	sequences map[uint64]sequenceWrapper
}

func (m sequence) nextId() sequenceWrapper {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	for {
		m.id++
		if m.id == 0 {
			m.id++
		}
		if _, has := m.sequences[m.id]; !has {
			break
		}
	}
	w := sequenceWrapper{
		id: m.id,
		messageCh: make(chan message),
		createdAt: time.Now(),
	}
	m.sequences[m.id] = w
	return w
}

func (m sequence) getId(id uint64) *sequenceWrapper {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	w, has := m.sequences[id]
	log.Println(">>>>>>>>>>w", w, has, id, m.sequences)
	if !has {
		return nil
	}
	delete(m.sequences, id)
	return &w
}

type sequenceWrapper struct {
	id uint64
	messageCh chan message
	createdAt time.Time
}

const (
	msgTypeSend = 0
	msgTypeAsk = 1
	msgTypeKill = 2
)
