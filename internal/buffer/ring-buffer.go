package buffer

import (
	"sync"

	"github.com/Gurveer1510/logaggregator/internal/model"
)

type RingBuffer struct {
	Data       []*model.LogEntry
	size       int
	lastInsert int
	nextRead   int
	mu         sync.RWMutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		Data:       make([]*model.LogEntry, size),
		size:       size,
		lastInsert: -1, // identifies that no inserts are occured
	}
}

func (r *RingBuffer) Insert(input model.LogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastInsert = (r.lastInsert + 1) % r.size
	r.Data[r.lastInsert] = &input

	if r.nextRead == r.lastInsert {
		r.nextRead = (r.nextRead + 1) % r.size
	}
}

func (r *RingBuffer) Query() []*model.LogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	output := []*model.LogEntry{}
	index := r.nextRead
	for {
		if r.Data[index] != nil {
			output = append(output, r.Data[index])
		}
		if index == r.lastInsert || r.lastInsert == -1 {
			break
		}
		index = (index + 1) % r.size
	}
	return output
}
