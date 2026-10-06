package handler

import (
	"encoding/json"
	"net/http"

	"server-monitor/backend/internal/collector"
	"server-monitor/backend/internal/store"
)

// ServerHandler handles /api/v1/server and /api/v1/status.
type ServerHandler struct {
	store store.MetricsStore
}

// NewServerHandler creates a new ServerHandler.
func NewServerHandler(s store.MetricsStore) *ServerHandler {
	return &ServerHandler{store: s}
}

// GetInfo returns static host identity and hardware profile.
func (h *ServerHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	info := h.store.GetHostInfo()
	info.Uptime = collector.GetUptime()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(info)
}

// GetStatus returns the current server health status and timestamp.
func (h *ServerHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	latest, ok := h.store.GetLatest()
	status := "healthy"
	uptime := collector.GetUptime()
	var ts int64

	if ok {
		status = latest.Status
		ts = latest.Timestamp
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"timestamp": ts,
		"uptime":    uptime,
	})
}
