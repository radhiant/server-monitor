package handler

import (
	"encoding/json"
	"net/http"

	"server-monitor/backend/internal/collector"
	"server-monitor/backend/internal/config"
	"server-monitor/backend/internal/model"
)

// HealthHandler handles /api/v1/health.
type HealthHandler struct {
	cfg *config.Config
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := model.HealthResponse{
		Status:   "ok",
		ServerID: h.cfg.ServerID,
		Uptime:   collector.GetUptime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
