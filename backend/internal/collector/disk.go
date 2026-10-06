package collector

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"

	"server-monitor/backend/internal/model"
)

// DiskCollector manages disk partition polling and delta I/O calculations.
type DiskCollector struct {
	mu           sync.Mutex
	lastIoTime   time.Time
	lastReadBytes  uint64
	lastWriteBytes uint64
	lastReadCount  uint64
	lastWriteCount uint64
	initialized  bool
}

// NewDiskCollector creates an initialized DiskCollector.
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

// CollectFilesystems gathers all mounted physical filesystem partition usages.
func (dc *DiskCollector) CollectFilesystems() []model.DiskPartition {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return []model.DiskPartition{}
	}

	seenMounts := make(map[string]bool)
	result := make([]model.DiskPartition, 0, len(partitions))

	for _, p := range partitions {
		if seenMounts[p.Mountpoint] {
			continue
		}
		// Ignore virtual filesystems
		if p.Fstype == "tmpfs" || p.Fstype == "devtmpfs" || p.Fstype == "squashfs" || p.Fstype == "overlay" {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage == nil || usage.Total == 0 {
			continue
		}

		seenMounts[p.Mountpoint] = true
		result = append(result, model.DiskPartition{
			Device:     p.Device,
			Mountpoint: p.Mountpoint,
			Fstype:     p.Fstype,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
			Usage:      usage.UsedPercent,
		})
	}

	return result
}

// CollectIO computes disk read/write delta rates per second.
func (dc *DiskCollector) CollectIO() model.DiskIoMetrics {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	now := time.Now()
	counters, err := disk.IOCounters()
	if err != nil {
		return model.DiskIoMetrics{}
	}

	var totalReadBytes, totalWriteBytes, totalReadCount, totalWriteCount uint64
	for _, c := range counters {
		totalReadBytes += c.ReadBytes
		totalWriteBytes += c.WriteBytes
		totalReadCount += c.ReadCount
		totalWriteCount += c.WriteCount
	}

	if !dc.initialized {
		dc.lastIoTime = now
		dc.lastReadBytes = totalReadBytes
		dc.lastWriteBytes = totalWriteBytes
		dc.lastReadCount = totalReadCount
		dc.lastWriteCount = totalWriteCount
		dc.initialized = true
		return model.DiskIoMetrics{}
	}

	deltaSec := now.Sub(dc.lastIoTime).Seconds()
	if deltaSec <= 0 {
		deltaSec = 1.0
	}

	var readRate, writeRate, readCountRate, writeCountRate float64
	if totalReadBytes >= dc.lastReadBytes {
		readRate = float64(totalReadBytes-dc.lastReadBytes) / deltaSec
	}
	if totalWriteBytes >= dc.lastWriteBytes {
		writeRate = float64(totalWriteBytes-dc.lastWriteBytes) / deltaSec
	}
	if totalReadCount >= dc.lastReadCount {
		readCountRate = float64(totalReadCount-dc.lastReadCount) / deltaSec
	}
	if totalWriteCount >= dc.lastWriteCount {
		writeCountRate = float64(totalWriteCount-dc.lastWriteCount) / deltaSec
	}

	dc.lastIoTime = now
	dc.lastReadBytes = totalReadBytes
	dc.lastWriteBytes = totalWriteBytes
	dc.lastReadCount = totalReadCount
	dc.lastWriteCount = totalWriteCount

	return model.DiskIoMetrics{
		ReadBytesSec:  readRate,
		WriteBytesSec: writeRate,
		ReadCountSec:  readCountRate,
		WriteCountSec: writeCountRate,
	}
}
