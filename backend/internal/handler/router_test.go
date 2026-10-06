package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server-monitor/backend/internal/config"
	"server-monitor/backend/internal/model"
	"server-monitor/backend/internal/store"
	"server-monitor/backend/internal/websocket"
)

func TestRouterHealthAndAuth(t *testing.T) {
	cfg := &config.Config{
		ServerID:    "test-srv",
		APIToken:    "secret123",
		CorsOrigins: []string{"*"},
		HistorySize: 60,
	}

	memStore := store.NewMemoryStore(60)
	hub := websocket.NewHub(func() any { return nil })
	router := NewRouter(cfg, memStore, hub)

	// 1. Health should be public (no auth required)
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", w.Code)
	}

	var healthResp model.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}
	if healthResp.ServerID != "test-srv" || healthResp.Status != "ok" {
		t.Fatalf("unexpected health response: %+v", healthResp)
	}

	// 2. Metrics without token should return 401 Unauthorized
	reqMetricsUnauth := httptest.NewRequest("GET", "/api/v1/metrics", nil)
	wMetricsUnauth := httptest.NewRecorder()
	router.ServeHTTP(wMetricsUnauth, reqMetricsUnauth)

	if wMetricsUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", wMetricsUnauth.Code)
	}

	// 3. Metrics with valid Bearer token
	memStore.PutLatest(model.ServerMetrics{
		Timestamp: 12345,
		ServerID:  "test-srv",
		Status:    "healthy",
	})

	reqMetricsAuth := httptest.NewRequest("GET", "/api/v1/metrics", nil)
	reqMetricsAuth.Header.Set("Authorization", "Bearer secret123")
	wMetricsAuth := httptest.NewRecorder()
	router.ServeHTTP(wMetricsAuth, reqMetricsAuth)

	if wMetricsAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", wMetricsAuth.Code, wMetricsAuth.Body.String())
	}

	// 4. Prometheus /metrics should be public
	reqProm := httptest.NewRequest("GET", "/metrics", nil)
	wProm := httptest.NewRecorder()
	router.ServeHTTP(wProm, reqProm)

	if wProm.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /metrics, got %d", wProm.Code)
	}
	if !strings.Contains(wProm.Body.String(), "server_uptime_seconds") {
		t.Fatalf("expected prometheus output to contain server_uptime_seconds, got %s", wProm.Body.String())
	}
}
