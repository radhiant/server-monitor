package store

import (
	"server-monitor/backend/internal/model"
)

// MetricsStore abstracts in-memory metrics storage without relying on any external database.
type MetricsStore interface {
	// SetHostInfo stores the server's static host identity and hardware profile.
	SetHostInfo(info model.HostInfo)

	// GetHostInfo retrieves static host profile.
	GetHostInfo() model.HostInfo

	// PutLatest records the newest collected metric sample and updates rolling history.
	PutLatest(metrics model.ServerMetrics)

	// GetLatest returns the most recent metrics snapshot.
	GetLatest() (model.ServerMetrics, bool)

	// GetHistory returns samples up to the requested count or time range.
	GetHistory(limit int) []model.ServerMetrics

	// GetFilesystems returns the latest filesystem partitions.
	GetFilesystems() []model.DiskPartition

	// GetProcesses returns the latest top process list.
	GetProcesses() []model.ProcessInfo

	// GetInterfaces returns the latest network interface states.
	GetInterfaces() []model.NetworkInterface
}
