package collector

import (
	"runtime"

	"github.com/shirou/gopsutil/v4/cpu"

	"server-monitor/backend/internal/model"
)

// CollectCPU gathers CPU metrics.
func CollectCPU() model.CpuMetrics {
	cores, _ := cpu.Counts(true)
	if cores <= 0 {
		cores = runtime.NumCPU()
	}

	// 0 duration returns metrics since last call
	totalPercents, err := cpu.Percent(0, false)
	totalUsage := 0.0
	if err == nil && len(totalPercents) > 0 {
		totalUsage = totalPercents[0]
	}

	perCorePercents, err := cpu.Percent(0, true)
	if err != nil || len(perCorePercents) == 0 {
		perCorePercents = make([]float64, cores)
		for i := range perCorePercents {
			perCorePercents[i] = totalUsage
		}
	}

	var freq float64
	var modelName string
	info, err := cpu.Info()
	if err == nil && len(info) > 0 {
		freq = info[0].Mhz
		modelName = info[0].ModelName
	}

	return model.CpuMetrics{
		Usage:     totalUsage,
		Cores:     cores,
		PerCore:   perCorePercents,
		Frequency: freq,
		ModelName: modelName,
	}
}
