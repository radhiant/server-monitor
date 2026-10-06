package store

import (
	"sync"

	"server-monitor/backend/internal/model"
)

// MemoryStore implements MetricsStore in-memory with zero external dependencies.
type MemoryStore struct {
	mu          sync.RWMutex
	hostInfo    model.HostInfo
	ringBuffer  *RingBuffer
	processes   []model.ProcessInfo
	filesystems []model.DiskPartition
	interfaces  []model.NetworkInterface
}

// NewMemoryStore initializes a new MemoryStore.
func NewMemoryStore(historyCapacity int) *MemoryStore {
	return &MemoryStore{
		ringBuffer:  NewRingBuffer(historyCapacity),
		processes:   make([]model.ProcessInfo, 0),
		filesystems: make([]model.DiskPartition, 0),
		interfaces:  make([]model.NetworkInterface, 0),
	}
}

func (s *MemoryStore) SetHostInfo(info model.HostInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hostInfo = info
}

func (s *MemoryStore) GetHostInfo() model.HostInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hostInfo
}

func (s *MemoryStore) PutLatest(metrics model.ServerMetrics) {
	s.mu.Lock()
	s.filesystems = metrics.Filesystems
	s.processes = metrics.Processes
	s.interfaces = metrics.Network
	s.mu.Unlock()

	s.ringBuffer.Push(metrics)
}

func (s *MemoryStore) GetLatest() (model.ServerMetrics, bool) {
	return s.ringBuffer.Latest()
}

func (s *MemoryStore) GetHistory(limit int) []model.ServerMetrics {
	return s.ringBuffer.GetLastN(limit)
}

func (s *MemoryStore) GetFilesystems() []model.DiskPartition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]model.DiskPartition, len(s.filesystems))
	copy(res, s.filesystems)
	return res
}

func (s *MemoryStore) GetProcesses() []model.ProcessInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]model.ProcessInfo, len(s.processes))
	copy(res, s.processes)
	return res
}

func (s *MemoryStore) GetInterfaces() []model.NetworkInterface {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]model.NetworkInterface, len(s.interfaces))
	copy(res, s.interfaces)
	return res
}
