package store

import (
	"sync"

	"server-monitor/backend/internal/model"
)

// RingBuffer provides a thread-safe circular buffer for rolling metric history.
type RingBuffer struct {
	mu       sync.RWMutex
	capacity int
	data     []model.ServerMetrics
	head     int
	size     int
}

// NewRingBuffer initializes a ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 60
	}
	return &RingBuffer{
		capacity: capacity,
		data:     make([]model.ServerMetrics, capacity),
		head:     0,
		size:     0,
	}
}

// Push adds a new metrics snapshot into the circular buffer.
func (r *RingBuffer) Push(item model.ServerMetrics) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[r.head] = item
	r.head = (r.head + 1) % r.capacity
	if r.size < r.capacity {
		r.size++
	}
}

// GetAll returns all items in chronological order (oldest to newest).
func (r *RingBuffer) GetAll() []model.ServerMetrics {
	return r.GetLastN(r.capacity)
}

// GetLastN returns the most recent N items in chronological order.
func (r *RingBuffer) GetLastN(n int) []model.ServerMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.size == 0 || n <= 0 {
		return []model.ServerMetrics{}
	}

	count := n
	if count > r.size {
		count = r.size
	}

	result := make([]model.ServerMetrics, count)
	startIdx := (r.head - count + r.capacity) % r.capacity

	for i := 0; i < count; i++ {
		idx := (startIdx + i) % r.capacity
		result[i] = r.data[idx]
	}

	return result
}

// Latest returns the most recently inserted sample.
func (r *RingBuffer) Latest() (model.ServerMetrics, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.size == 0 {
		return model.ServerMetrics{}, false
	}

	latestIdx := (r.head - 1 + r.capacity) % r.capacity
	return r.data[latestIdx], true
}
