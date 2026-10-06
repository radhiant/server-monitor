package collector

import (
	"github.com/shirou/gopsutil/v4/mem"

	"server-monitor/backend/internal/model"
)

// CollectMemory gathers virtual memory and swap statistics.
func CollectMemory() model.MemoryMetrics {
	var metrics model.MemoryMetrics

	vMem, err := mem.VirtualMemory()
	if err == nil && vMem != nil {
		metrics.Total = vMem.Total
		metrics.Used = vMem.Used
		metrics.Available = vMem.Available
		metrics.Free = vMem.Free
		metrics.Usage = vMem.UsedPercent
		metrics.Cached = vMem.Cached
		metrics.Buffers = vMem.Buffers
	}

	swap, err := mem.SwapMemory()
	if err == nil && swap != nil {
		metrics.SwapTotal = swap.Total
		metrics.SwapUsed = swap.Used
		metrics.SwapFree = swap.Free
		metrics.SwapUsage = swap.UsedPercent
	}

	return metrics
}
