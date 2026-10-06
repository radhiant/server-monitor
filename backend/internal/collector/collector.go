package collector

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"server-monitor/backend/internal/model"
	"server-monitor/backend/internal/notify"
	"server-monitor/backend/internal/store"
)

// Broadcaster is an interface to decouple WebSocket broadcasting from collector logic.
type Broadcaster interface {
	Broadcast(msg model.WebSocketMessage)
}

// Engine coordinates system metric collection at defined intervals.
type Engine struct {
	serverID      string
	interval      time.Duration
	store         store.MetricsStore
	broadcaster   Broadcaster
	notifier      notify.Notifier
	diskCollector *DiskCollector
	netCollector  *NetworkCollector
	hostInfo      model.HostInfo
}

// NewEngine initializes the metrics collector engine.
func NewEngine(serverID string, interval time.Duration, s store.MetricsStore, b Broadcaster, n notify.Notifier) *Engine {
	return &Engine{
		serverID:      serverID,
		interval:      interval,
		store:         s,
		broadcaster:   b,
		notifier:      n,
		diskCollector: NewDiskCollector(),
		netCollector:  NewNetworkCollector(),
	}
}

// Start runs the centralized collection loop until the context is canceled.
func (e *Engine) Start(ctx context.Context) {
	log.Info().
		Str("server_id", e.serverID).
		Dur("interval", e.interval).
		Msg("starting centralized metrics collector")

	// Store initial host profile
	e.hostInfo = CollectHostInfo(e.serverID)
	e.store.SetHostInfo(e.hostInfo)

	// Prime collectors
	e.diskCollector.CollectIO()
	e.netCollector.Collect()

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	// Initial immediate collection tick
	e.collectTick()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("metrics collector stopped")
			return
		case <-ticker.C:
			e.collectTick()
		}
	}
}

func (e *Engine) collectTick() {
	now := time.Now().Unix()
	uptime := GetUptime()

	cpuMetrics := CollectCPU()
	memMetrics := CollectMemory()
	fsMetrics := e.diskCollector.CollectFilesystems()
	diskIoMetrics := e.diskCollector.CollectIO()
	netMetrics := e.netCollector.Collect()
	loadMetrics := CollectLoad(cpuMetrics.Cores)
	procMetrics := CollectProcesses(20)

	// Determine overall server health status
	status := "healthy"
	if cpuMetrics.Usage >= 95 || memMetrics.Usage >= 95 || loadMetrics.Status == "critical" {
		status = "critical"
	} else if cpuMetrics.Usage >= 80 || memMetrics.Usage >= 80 || loadMetrics.Status == "warning" {
		status = "warning"
	}

	// Refresh uptime on host profile
	currentHost := e.hostInfo
	currentHost.Uptime = uptime

	snapshot := model.ServerMetrics{
		Timestamp:   now,
		ServerID:    e.serverID,
		Status:      status,
		Uptime:      uptime,
		Host:        &currentHost,
		Cpu:         cpuMetrics,
		Memory:      memMetrics,
		Filesystems: fsMetrics,
		DiskIO:      diskIoMetrics,
		Network:     netMetrics,
		Load:        loadMetrics,
		Processes:   procMetrics,
	}

	// Update local in-memory storage
	e.store.PutLatest(snapshot)

	// Broadcast to WebSocket clients if broadcaster is present
	if e.broadcaster != nil {
		e.broadcaster.Broadcast(model.WebSocketMessage{
			Type:      "metrics",
			Timestamp: now,
			Data:      snapshot,
		})
	}

	// Check alerts via notifier hook
	if e.notifier != nil {
		e.notifier.Check(&snapshot)
	}
}
