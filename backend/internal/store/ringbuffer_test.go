package store

import (
	"testing"

	"server-monitor/backend/internal/model"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer(5)

	// Test empty
	if _, ok := rb.Latest(); ok {
		t.Fatal("expected empty ring buffer")
	}
	if len(rb.GetAll()) != 0 {
		t.Fatal("expected 0 items")
	}

	// Push 3 items
	for i := int64(1); i <= 3; i++ {
		rb.Push(model.ServerMetrics{Timestamp: i})
	}

	if latest, ok := rb.Latest(); !ok || latest.Timestamp != 3 {
		t.Fatalf("expected latest timestamp 3, got %v", latest.Timestamp)
	}

	all := rb.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 items, got %d", len(all))
	}
	if all[0].Timestamp != 1 || all[2].Timestamp != 3 {
		t.Fatalf("unexpected ordering: %+v", all)
	}

	// Push items to overflow capacity (push 4, 5, 6, 7)
	for i := int64(4); i <= 7; i++ {
		rb.Push(model.ServerMetrics{Timestamp: i})
	}

	// Buffer should now contain [3, 4, 5, 6, 7]
	all = rb.GetAll()
	if len(all) != 5 {
		t.Fatalf("expected 5 items, got %d", len(all))
	}
	if all[0].Timestamp != 3 || all[4].Timestamp != 7 {
		t.Fatalf("expected [3,4,5,6,7], got oldest %d, newest %d", all[0].Timestamp, all[4].Timestamp)
	}

	// Test GetLastN
	last3 := rb.GetLastN(3)
	if len(last3) != 3 {
		t.Fatalf("expected 3 items, got %d", len(last3))
	}
	if last3[0].Timestamp != 5 || last3[2].Timestamp != 7 {
		t.Fatalf("expected [5,6,7], got oldest %d, newest %d", last3[0].Timestamp, last3[2].Timestamp)
	}
}
