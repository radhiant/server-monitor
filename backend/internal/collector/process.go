package collector

import (
	"sort"

	"github.com/shirou/gopsutil/v4/process"

	"server-monitor/backend/internal/model"
)

// CollectProcesses retrieves and sorts top 20 processes by CPU consumption.
func CollectProcesses(limit int) []model.ProcessInfo {
	if limit <= 0 {
		limit = 20
	}

	procs, err := process.Processes()
	if err != nil {
		return []model.ProcessInfo{}
	}

	results := make([]model.ProcessInfo, 0, len(procs))

	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}

		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		threads, _ := p.NumThreads()
		username, _ := p.Username()
		statusSlice, _ := p.Status()
		status := "running"
		if len(statusSlice) > 0 {
			status = statusSlice[0]
		}

		var rss, vms uint64
		if memInfo != nil {
			rss = memInfo.RSS
			vms = memInfo.VMS
		}

		results = append(results, model.ProcessInfo{
			PID:           p.Pid,
			Name:          name,
			CpuPercent:    cpuPct,
			MemoryPercent: memPct,
			MemoryRss:     rss,
			MemoryVms:     vms,
			NumThreads:    threads,
			Username:      username,
			Status:        status,
		})
	}

	// Sort descending by CPU percent
	sort.Slice(results, func(i, j int) bool {
		if results[i].CpuPercent == results[j].CpuPercent {
			return results[i].MemoryPercent > results[j].MemoryPercent
		}
		return results[i].CpuPercent > results[j].CpuPercent
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}
