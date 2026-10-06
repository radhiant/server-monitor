package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"server-monitor/backend/internal/config"
	"server-monitor/backend/internal/model"
)

func TestServerWebhookNotifier(t *testing.T) {
	var mu sync.Mutex
	var receivedPayloads []webhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p webhookPayload
		_ = json.Unmarshal(body, &p)

		mu.Lock()
		receivedPayloads = append(receivedPayloads, p)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		WebhookEnabled:     true,
		WebhookURL:         server.URL,
		WebhookChannel:     "both",
		WebhookChatID:      "-1000000000",
		WebhookWARecipient: "620000000000",
		CaptureURL:         "http://localhost:8080/",
		AlertCpuThreshold:  95.0,
		AlertMemThreshold:  95.0,
		AlertDebounceTicks: 3,
		AlertCooldown:      15 * time.Minute,
	}

	notifier := NewWebhookNotifier(cfg)

	critMetrics := &model.ServerMetrics{
		ServerID: "test-server",
		Status:   "critical",
		Cpu:      model.CpuMetrics{Usage: 96.5},
		Memory:   model.MemoryMetrics{Usage: 92.0},
		Load:     model.LoadMetrics{Load1: 8.5, Load5: 6.2, Load15: 4.1},
		Processes: []model.ProcessInfo{
			{PID: 1234, Name: "stress", CpuPercent: 90.0, MemoryPercent: 10.0},
		},
	}

	// 1st tick critical (should debounce, no alert yet)
	notifier.Check(critMetrics)
	// 2nd tick critical (should debounce, no alert yet)
	notifier.Check(critMetrics)

	mu.Lock()
	if len(receivedPayloads) != 0 {
		t.Fatalf("expected 0 alerts due to debounce, got %d", len(receivedPayloads))
	}
	mu.Unlock()

	// 3rd tick critical -> fires alert!
	notifier.Check(critMetrics)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if len(receivedPayloads) != 1 {
		t.Fatalf("expected 1 alert fired, got %d", len(receivedPayloads))
	}
	if receivedPayloads[0].URL != "http://localhost:8080/" {
		t.Errorf("unexpected capture URL: %s", receivedPayloads[0].URL)
	}
	mu.Unlock()

	// Healthy ticks (5 ticks required for recovery)
	healthyMetrics := &model.ServerMetrics{
		ServerID: "test-server",
		Status:   "healthy",
		Cpu:      model.CpuMetrics{Usage: 25.0},
		Memory:   model.MemoryMetrics{Usage: 30.0},
	}
	for i := 0; i < 5; i++ {
		notifier.Check(healthyMetrics)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if len(receivedPayloads) != 2 {
		t.Fatalf("expected 2 alerts (1 crit + 1 recovery), got %d", len(receivedPayloads))
	}
	mu.Unlock()
}
