package collector

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"

	"server-monitor/backend/internal/model"
)

// CollectHostInfo gathers system hardware and OS identity.
func CollectHostInfo(serverID string) model.HostInfo {
	info, err := host.Info()
	var (
		hostname    = serverID
		osName      = runtime.GOOS
		platform    = runtime.GOOS
		platFamily  = ""
		platVersion = ""
		kernelVer   = ""
		kernelArch  = runtime.GOARCH
		uptime      uint64
		bootTime    uint64
	)

	if err == nil && info != nil {
		if info.Hostname != "" {
			hostname = info.Hostname
		}
		osName = info.OS
		platform = info.Platform
		platFamily = info.PlatformFamily
		platVersion = info.PlatformVersion
		kernelVer = info.KernelVersion
		kernelArch = info.KernelArch
		uptime = info.Uptime
		bootTime = info.BootTime
	}

	cpuModel := ""
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	cores, _ := cpu.Counts(true)
	if cores <= 0 {
		cores = runtime.NumCPU()
	}

	var totalMem uint64
	vMem, err := mem.VirtualMemory()
	if err == nil && vMem != nil {
		totalMem = vMem.Total
	}

	return model.HostInfo{
		ServerID:        serverID,
		Hostname:        hostname,
		OS:              osName,
		Platform:        platform,
		PlatformFamily:  platFamily,
		PlatformVersion: platVersion,
		KernelVersion:   kernelVer,
		KernelArch:      kernelArch,
		Uptime:          uptime,
		BootTime:        bootTime,
		CpuModel:        cpuModel,
		CpuCores:        cores,
		TotalMemory:     totalMem,
	}
}

// GetUptime retrieves current uptime in seconds.
func GetUptime() uint64 {
	uptime, err := host.Uptime()
	if err != nil {
		return uint64(time.Now().Unix())
	}
	return uptime
}
