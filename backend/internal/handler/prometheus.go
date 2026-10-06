package handler

import (
	"fmt"
	"net/http"
	"strings"

	"server-monitor/backend/internal/config"
	"server-monitor/backend/internal/store"
)

// PrometheusHandler exports server metrics in standard Prometheus exposition format.
type PrometheusHandler struct {
	store store.MetricsStore
	cfg   *config.Config
}

// NewPrometheusHandler constructs a new PrometheusHandler.
func NewPrometheusHandler(s store.MetricsStore, cfg *config.Config) *PrometheusHandler {
	return &PrometheusHandler{
		store: s,
		cfg:   cfg,
	}
}

// escapeLabel escapes Prometheus label values.
func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

// ServeHTTP writes the Prometheus metrics formatted output.
func (h *PrometheusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	latest, ok := h.store.GetLatest()
	if !ok {
		http.Error(w, "Metrics collection is still warming up", http.StatusServiceUnavailable)
		return
	}

	host := h.store.GetHostInfo()
	serverID := escapeLabel(latest.ServerID)
	if serverID == "" && h.cfg != nil {
		serverID = escapeLabel(h.cfg.ServerID)
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder

	// Uptime & Status
	b.WriteString("# HELP server_uptime_seconds System uptime in seconds\n")
	b.WriteString("# TYPE server_uptime_seconds gauge\n")
	b.WriteString(fmt.Sprintf("server_uptime_seconds{server_id=\"%s\"} %d\n", serverID, latest.Uptime))

	b.WriteString("# HELP server_status System health status (1 for active status)\n")
	b.WriteString("# TYPE server_status gauge\n")
	for _, st := range []string{"healthy", "warning", "critical"} {
		val := 0
		if latest.Status == st {
			val = 1
		}
		b.WriteString(fmt.Sprintf("server_status{server_id=\"%s\",status=\"%s\"} %d\n", serverID, st, val))
	}

	// Host Info
	b.WriteString("# HELP server_info Static host and hardware metadata\n")
	b.WriteString("# TYPE server_info gauge\n")
	b.WriteString(fmt.Sprintf("server_info{server_id=\"%s\",hostname=\"%s\",os=\"%s\",platform=\"%s\",platform_version=\"%s\",kernel_version=\"%s\",kernel_arch=\"%s\"} 1\n",
		serverID,
		escapeLabel(host.Hostname),
		escapeLabel(host.OS),
		escapeLabel(host.Platform),
		escapeLabel(host.PlatformVersion),
		escapeLabel(host.KernelVersion),
		escapeLabel(host.KernelArch),
	))

	// CPU Metrics
	b.WriteString("# HELP server_cpu_usage_percent Overall CPU usage percentage\n")
	b.WriteString("# TYPE server_cpu_usage_percent gauge\n")
	b.WriteString(fmt.Sprintf("server_cpu_usage_percent{server_id=\"%s\"} %.2f\n", serverID, latest.Cpu.Usage))

	b.WriteString("# HELP server_cpu_cores Number of CPU cores\n")
	b.WriteString("# TYPE server_cpu_cores gauge\n")
	b.WriteString(fmt.Sprintf("server_cpu_cores{server_id=\"%s\"} %d\n", serverID, latest.Cpu.Cores))

	if len(latest.Cpu.PerCore) > 0 {
		b.WriteString("# HELP server_cpu_core_usage_percent Per-core CPU usage percentage\n")
		b.WriteString("# TYPE server_cpu_core_usage_percent gauge\n")
		for i, usage := range latest.Cpu.PerCore {
			b.WriteString(fmt.Sprintf("server_cpu_core_usage_percent{server_id=\"%s\",core=\"%d\"} %.2f\n", serverID, i, usage))
		}
	}

	// Memory Metrics
	b.WriteString("# HELP server_memory_total_bytes Total system memory in bytes\n")
	b.WriteString("# TYPE server_memory_total_bytes gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_total_bytes{server_id=\"%s\"} %d\n", serverID, latest.Memory.Total))

	b.WriteString("# HELP server_memory_used_bytes Used system memory in bytes\n")
	b.WriteString("# TYPE server_memory_used_bytes gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_used_bytes{server_id=\"%s\"} %d\n", serverID, latest.Memory.Used))

	b.WriteString("# HELP server_memory_available_bytes Available system memory in bytes\n")
	b.WriteString("# TYPE server_memory_available_bytes gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_available_bytes{server_id=\"%s\"} %d\n", serverID, latest.Memory.Available))

	b.WriteString("# HELP server_memory_free_bytes Free system memory in bytes\n")
	b.WriteString("# TYPE server_memory_free_bytes gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_free_bytes{server_id=\"%s\"} %d\n", serverID, latest.Memory.Free))

	b.WriteString("# HELP server_memory_usage_percent Memory usage percentage\n")
	b.WriteString("# TYPE server_memory_usage_percent gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_usage_percent{server_id=\"%s\"} %.2f\n", serverID, latest.Memory.Usage))

	b.WriteString("# HELP server_memory_swap_total_bytes Total swap memory in bytes\n")
	b.WriteString("# TYPE server_memory_swap_total_bytes gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_swap_total_bytes{server_id=\"%s\"} %d\n", serverID, latest.Memory.SwapTotal))

	b.WriteString("# HELP server_memory_swap_used_bytes Used swap memory in bytes\n")
	b.WriteString("# TYPE server_memory_swap_used_bytes gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_swap_used_bytes{server_id=\"%s\"} %d\n", serverID, latest.Memory.SwapUsed))

	b.WriteString("# HELP server_memory_swap_usage_percent Swap memory usage percentage\n")
	b.WriteString("# TYPE server_memory_swap_usage_percent gauge\n")
	b.WriteString(fmt.Sprintf("server_memory_swap_usage_percent{server_id=\"%s\"} %.2f\n", serverID, latest.Memory.SwapUsage))

	// Disk Metrics
	if len(latest.Filesystems) > 0 {
		b.WriteString("# HELP server_disk_total_bytes Total disk partition space in bytes\n")
		b.WriteString("# TYPE server_disk_total_bytes gauge\n")
		for _, fs := range latest.Filesystems {
			b.WriteString(fmt.Sprintf("server_disk_total_bytes{server_id=\"%s\",mountpoint=\"%s\",device=\"%s\",fstype=\"%s\"} %d\n",
				serverID, escapeLabel(fs.Mountpoint), escapeLabel(fs.Device), escapeLabel(fs.Fstype), fs.Total))
		}

		b.WriteString("# HELP server_disk_used_bytes Used disk partition space in bytes\n")
		b.WriteString("# TYPE server_disk_used_bytes gauge\n")
		for _, fs := range latest.Filesystems {
			b.WriteString(fmt.Sprintf("server_disk_used_bytes{server_id=\"%s\",mountpoint=\"%s\",device=\"%s\",fstype=\"%s\"} %d\n",
				serverID, escapeLabel(fs.Mountpoint), escapeLabel(fs.Device), escapeLabel(fs.Fstype), fs.Used))
		}

		b.WriteString("# HELP server_disk_free_bytes Free disk partition space in bytes\n")
		b.WriteString("# TYPE server_disk_free_bytes gauge\n")
		for _, fs := range latest.Filesystems {
			b.WriteString(fmt.Sprintf("server_disk_free_bytes{server_id=\"%s\",mountpoint=\"%s\",device=\"%s\",fstype=\"%s\"} %d\n",
				serverID, escapeLabel(fs.Mountpoint), escapeLabel(fs.Device), escapeLabel(fs.Fstype), fs.Free))
		}

		b.WriteString("# HELP server_disk_usage_percent Disk partition usage percentage\n")
		b.WriteString("# TYPE server_disk_usage_percent gauge\n")
		for _, fs := range latest.Filesystems {
			b.WriteString(fmt.Sprintf("server_disk_usage_percent{server_id=\"%s\",mountpoint=\"%s\",device=\"%s\",fstype=\"%s\"} %.2f\n",
				serverID, escapeLabel(fs.Mountpoint), escapeLabel(fs.Device), escapeLabel(fs.Fstype), fs.Usage))
		}
	}

	// Disk I/O
	b.WriteString("# HELP server_disk_read_bytes_per_sec Disk read rate in bytes per second\n")
	b.WriteString("# TYPE server_disk_read_bytes_per_sec gauge\n")
	b.WriteString(fmt.Sprintf("server_disk_read_bytes_per_sec{server_id=\"%s\"} %.2f\n", serverID, latest.DiskIO.ReadBytesSec))

	b.WriteString("# HELP server_disk_write_bytes_per_sec Disk write rate in bytes per second\n")
	b.WriteString("# TYPE server_disk_write_bytes_per_sec gauge\n")
	b.WriteString(fmt.Sprintf("server_disk_write_bytes_per_sec{server_id=\"%s\"} %.2f\n", serverID, latest.DiskIO.WriteBytesSec))

	// Network Metrics
	if len(latest.Network) > 0 {
		b.WriteString("# HELP server_network_rx_bytes_total Total received bytes on network interface\n")
		b.WriteString("# TYPE server_network_rx_bytes_total counter\n")
		for _, iface := range latest.Network {
			b.WriteString(fmt.Sprintf("server_network_rx_bytes_total{server_id=\"%s\",interface=\"%s\"} %d\n",
				serverID, escapeLabel(iface.Name), iface.TotalRxBytes))
		}

		b.WriteString("# HELP server_network_tx_bytes_total Total transmitted bytes on network interface\n")
		b.WriteString("# TYPE server_network_tx_bytes_total counter\n")
		for _, iface := range latest.Network {
			b.WriteString(fmt.Sprintf("server_network_tx_bytes_total{server_id=\"%s\",interface=\"%s\"} %d\n",
				serverID, escapeLabel(iface.Name), iface.TotalTxBytes))
		}

		b.WriteString("# HELP server_network_rx_bytes_per_sec Network receive rate in bytes per second\n")
		b.WriteString("# TYPE server_network_rx_bytes_per_sec gauge\n")
		for _, iface := range latest.Network {
			b.WriteString(fmt.Sprintf("server_network_rx_bytes_per_sec{server_id=\"%s\",interface=\"%s\"} %.2f\n",
				serverID, escapeLabel(iface.Name), iface.RxBytesSec))
		}

		b.WriteString("# HELP server_network_tx_bytes_per_sec Network transmit rate in bytes per second\n")
		b.WriteString("# TYPE server_network_tx_bytes_per_sec gauge\n")
		for _, iface := range latest.Network {
			b.WriteString(fmt.Sprintf("server_network_tx_bytes_per_sec{server_id=\"%s\",interface=\"%s\"} %.2f\n",
				serverID, escapeLabel(iface.Name), iface.TxBytesSec))
		}
	}

	// Load Metrics
	b.WriteString("# HELP server_load1 System 1-minute load average\n")
	b.WriteString("# TYPE server_load1 gauge\n")
	b.WriteString(fmt.Sprintf("server_load1{server_id=\"%s\"} %.2f\n", serverID, latest.Load.Load1))

	b.WriteString("# HELP server_load5 System 5-minute load average\n")
	b.WriteString("# TYPE server_load5 gauge\n")
	b.WriteString(fmt.Sprintf("server_load5{server_id=\"%s\"} %.2f\n", serverID, latest.Load.Load5))

	b.WriteString("# HELP server_load15 System 15-minute load average\n")
	b.WriteString("# TYPE server_load15 gauge\n")
	b.WriteString(fmt.Sprintf("server_load15{server_id=\"%s\"} %.2f\n", serverID, latest.Load.Load15))

	// Processes Count
	if len(latest.Processes) > 0 {
		b.WriteString("# HELP server_processes_total Top tracked processes count\n")
		b.WriteString("# TYPE server_processes_total gauge\n")
		b.WriteString(fmt.Sprintf("server_processes_total{server_id=\"%s\"} %d\n", serverID, len(latest.Processes)))

		b.WriteString("# HELP server_process_cpu_percent Top process CPU percent\n")
		b.WriteString("# TYPE server_process_cpu_percent gauge\n")
		for _, p := range latest.Processes {
			b.WriteString(fmt.Sprintf("server_process_cpu_percent{server_id=\"%s\",pid=\"%d\",name=\"%s\",user=\"%s\"} %.2f\n",
				serverID, p.PID, escapeLabel(p.Name), escapeLabel(p.Username), p.CpuPercent))
		}
	}

	_, _ = w.Write([]byte(b.String()))
}
