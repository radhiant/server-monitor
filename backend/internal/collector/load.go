package collector

import (
	"github.com/shirou/gopsutil/v4/load"

	"server-monitor/backend/internal/model"
)

// CollectLoad gathers 1m, 5m, 15m load averages and computes health status relative to core count.
func CollectLoad(cores int) model.LoadMetrics {
	if cores <= 0 {
		cores = 1
	}

	avg, err := load.Avg()
	if err != nil || avg == nil {
		// Windows fallback: return zeroes
		return model.LoadMetrics{
			Load1:  0,
			Load5:  0,
			Load15: 0,
			Status: "healthy",
		}
	}

	// Calculate status contextual to CPU core count
	// E.g. Load1 / Cores > 1.0 (Warning), > 2.0 (Critical)
	ratio := avg.Load1 / float64(cores)
	status := "healthy"
	if ratio >= 1.5 {
		status = "critical"
	} else if ratio >= 0.9 {
		status = "warning"
	}

	return model.LoadMetrics{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
		Status: status,
	}
}
