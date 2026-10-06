package model

// HostInfo holds static and semi-static host system properties.
type HostInfo struct {
	ServerID        string `json:"server_id"`
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	KernelArch      string `json:"kernel_arch"`
	Uptime          uint64 `json:"uptime"`
	BootTime        uint64 `json:"boot_time"`
	CpuModel        string `json:"cpu_model,omitempty"`
	CpuCores        int    `json:"cpu_cores"`
	TotalMemory     uint64 `json:"total_memory"`
}

// CpuMetrics holds realtime CPU utilization metrics.
type CpuMetrics struct {
	Usage     float64   `json:"usage"`
	Cores     int       `json:"cores"`
	PerCore   []float64 `json:"per_core"`
	Frequency float64   `json:"frequency_mhz,omitempty"`
	ModelName string    `json:"model_name,omitempty"`
}

// MemoryMetrics holds RAM and Swap usage figures in bytes and percentages.
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	Free        uint64  `json:"free"`
	Usage       float64 `json:"usage"`
	Cached      uint64  `json:"cached"`
	Buffers     uint64  `json:"buffers"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapFree    uint64  `json:"swap_free"`
	SwapUsage   float64 `json:"swap_usage"`
}

// DiskPartition describes mounted partition usage.
type DiskPartition struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Fstype     string  `json:"fstype"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	Usage      float64 `json:"usage"`
}

// DiskIoMetrics describes realtime disk I/O rates.
type DiskIoMetrics struct {
	ReadBytesSec   float64 `json:"read_bytes_sec"`
	WriteBytesSec  float64 `json:"write_bytes_sec"`
	ReadCountSec   float64 `json:"read_count_sec"`
	WriteCountSec  float64 `json:"write_count_sec"`
}

// NetworkInterface holds per-interface transmission rates and counters.
type NetworkInterface struct {
	Name         string  `json:"name"`
	RxBytesSec   float64 `json:"rx_bytes_sec"`
	TxBytesSec   float64 `json:"tx_bytes_sec"`
	RxPacketsSec float64 `json:"rx_packets_sec"`
	TxPacketsSec float64 `json:"tx_packets_sec"`
	TotalRxBytes uint64  `json:"total_rx_bytes"`
	TotalTxBytes uint64  `json:"total_tx_bytes"`
}

// LoadMetrics holds unix 1, 5, 15 minute load averages.
type LoadMetrics struct {
	Load1  float64 `json:"1m"`
	Load5  float64 `json:"5m"`
	Load15 float64 `json:"15m"`
	Status string  `json:"status"` // healthy, warning, critical
}

// ProcessInfo holds top process statistics.
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CpuPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	MemoryRss     uint64  `json:"memory_rss"`
	MemoryVms     uint64  `json:"memory_vms"`
	NumThreads    int32   `json:"num_threads"`
	Username      string  `json:"username"`
	Status        string  `json:"status"`
}

// ServerMetrics is the complete snapshot produced on every collection cycle.
type ServerMetrics struct {
	Timestamp   int64              `json:"timestamp"`
	ServerID    string             `json:"server_id"`
	Status      string             `json:"status"` // healthy, warning, critical
	Uptime      uint64             `json:"uptime"`
	Host        *HostInfo          `json:"host,omitempty"`
	Cpu         CpuMetrics         `json:"cpu"`
	Memory      MemoryMetrics      `json:"memory"`
	Filesystems []DiskPartition    `json:"filesystems"`
	DiskIO      DiskIoMetrics      `json:"disk_io"`
	Network     []NetworkInterface `json:"network"`
	Load        LoadMetrics        `json:"load"`
	Processes   []ProcessInfo      `json:"processes,omitempty"`
}

// WebSocketMessage wraps typed realtime broadcasts.
type WebSocketMessage struct {
	Type      string      `json:"type"` // "metrics", "processes", "server_status", "error", "pong"
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// HealthResponse returned by /api/v1/health.
type HealthResponse struct {
	Status   string `json:"status"`
	ServerID string `json:"server_id"`
	Uptime   uint64 `json:"uptime"`
}

// ErrorDetail payload for standardized error responses.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse returned upon any API or WS error.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
