package hub

import (
	"sync"

	"github.com/Gurveer1510/logaggregator/internal/model"
)

type Hub struct {
	subscribers []chan model.LogEntry
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make([]chan model.LogEntry, 0),
	}
}

func (h *Hub) Subscribe(ch chan model.LogEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscribers = append(h.subscribers, ch)
}
func (h *Hub) Unsubscribe(ch chan model.LogEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	indexToRemove := 0
	found := false
	for index, v := range h.subscribers {
		if v == ch {
			indexToRemove = index
			found = true
			break
		}
	}

	if !found {
		return // Channel not found, nothing to remove
	}

	h.subscribers[indexToRemove] = h.subscribers[len(h.subscribers)-1]
	h.subscribers = h.subscribers[:len(h.subscribers)-1]
	close(ch) // Close the channel to signal the subscriber
}

func (h *Hub) Broadcast(logentry model.LogEntry) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, v := range h.subscribers {
		select {
		case v <- logentry:
		default:
			
		}
	}
}