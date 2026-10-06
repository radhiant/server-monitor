package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server-monitor/backend/internal/model"
	"server-monitor/backend/internal/store"
)

// MetricsHandler handles metrics, history, filesystems, and network interface endpoints.
type MetricsHandler struct {
	store store.MetricsStore
}

// NewMetricsHandler creates a new MetricsHandler.
func NewMetricsHandler(s store.MetricsStore) *MetricsHandler {
	return &MetricsHandler{store: s}
}

// GetLatest returns the most recent collected metrics snapshot.
func (h *MetricsHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	latest, ok := h.store.GetLatest()
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "METRICS_NOT_READY",
				Message: "Metrics collection is still warming up",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(latest)
}

// GetHistory returns the rolling metrics history.
func (h *MetricsHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	limit := 60
	rangeQuery := r.URL.Query().Get("range")
	switch rangeQuery {
	case "1m":
		limit = 60
	case "5m":
		limit = 300
	case "15m":
		limit = 900
	default:
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}
	}

	history := h.store.GetHistory(limit)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"count":   len(history),
		"samples": history,
	})
}

// GetFilesystems returns the latest filesystem usage breakdown.
func (h *MetricsHandler) GetFilesystems(w http.ResponseWriter, r *http.Request) {
	fs := h.store.GetFilesystems()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(fs)
}

// GetInterfaces returns the latest network interface states.
func (h *MetricsHandler) GetInterfaces(w http.ResponseWriter, r *http.Request) {
	ifaces := h.store.GetInterfaces()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ifaces)
}
