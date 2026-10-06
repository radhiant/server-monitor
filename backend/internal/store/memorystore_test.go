package store

import (
	"testing"

	"server-monitor/backend/internal/model"
)

func TestMemoryStore(t *testing.T) {
	ms := NewMemoryStore(10)

	// Set & Get HostInfo
	host := model.HostInfo{
		ServerID: "test-server-01",
		Hostname: "node1.local",
		OS:       "linux",
		CpuCores: 8,
	}
	ms.SetHostInfo(host)

	gotHost := ms.GetHostInfo()
	if gotHost.ServerID != "test-server-01" || gotHost.CpuCores != 8 {
		t.Fatalf("unexpected host info: %+v", gotHost)
	}

	// Put and retrieve metrics
	sample := model.ServerMetrics{
		Timestamp: 1000,
		ServerID:  "test-server-01",
		Status:    "healthy",
		Filesystems: []model.DiskPartition{
			{Mountpoint: "/", Total: 100, Used: 20},
		},
		Processes: []model.ProcessInfo{
			{PID: 1234, Name: "nginx", CpuPercent: 2.5},
		},
		Network: []model.NetworkInterface{
			{Name: "eth0", RxBytesSec: 1024},
		},
	}

	ms.PutLatest(sample)

	latest, ok := ms.GetLatest()
	if !ok || latest.Timestamp != 1000 {
		t.Fatalf("expected latest timestamp 1000, got %+v", latest)
	}

	fs := ms.GetFilesystems()
	if len(fs) != 1 || fs[0].Mountpoint != "/" {
		t.Fatalf("expected 1 filesystem, got %+v", fs)
	}

	procs := ms.GetProcesses()
	if len(procs) != 1 || procs[0].Name != "nginx" {
		t.Fatalf("expected 1 process, got %+v", procs)
	}

	ifaces := ms.GetInterfaces()
	if len(ifaces) != 1 || ifaces[0].Name != "eth0" {
		t.Fatalf("expected 1 interface, got %+v", ifaces)
	}
}
