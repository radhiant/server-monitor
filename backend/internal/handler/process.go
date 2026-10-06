package handler

import (
	"encoding/json"
	"net/http"

	"server-monitor/backend/internal/store"
)

// ProcessHandler handles /api/v1/processes.
type ProcessHandler struct {
	store store.MetricsStore
}

// NewProcessHandler creates a new ProcessHandler.
func NewProcessHandler(s store.MetricsStore) *ProcessHandler {
	return &ProcessHandler{store: s}
}

func (h *ProcessHandler) GetProcesses(w http.ResponseWriter, r *http.Request) {
	procs := h.store.GetProcesses()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(procs)
}
