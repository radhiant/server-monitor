package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"server-monitor/backend/internal/config"
	"server-monitor/backend/internal/model"
)

// Notifier defines the interface for server health alerting.
type Notifier interface {
	Check(metrics *model.ServerMetrics)
}

// NopNotifier does nothing.
type NopNotifier struct{}

func (NopNotifier) Check(*model.ServerMetrics) {}

// WebhookNotifier sends server critical and recovery alerts to an n8n webhook.
type WebhookNotifier struct {
	cfg        *config.Config
	httpClient *http.Client

	mu            sync.Mutex
	critCount     int
	healthyCount  int
	isAlerting    bool
	alertedAt     time.Time
	lastAlertTime time.Time
}

// NewWebhookNotifier creates an initialized WebhookNotifier.
func NewWebhookNotifier(cfg *config.Config) *WebhookNotifier {
	return &WebhookNotifier{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type webhookPayload struct {
	URL         string `json:"url"`
	Channel     string `json:"channel"`
	Caption     string `json:"caption"`
	ChatID      string `json:"chatId,omitempty"`
	WARecipient string `json:"waRecipient,omitempty"`
}

func (w *WebhookNotifier) send(caption string) {
	if !w.cfg.WebhookEnabled || w.cfg.WebhookURL == "" {
		return
	}

	payload := webhookPayload{
		URL:         w.cfg.CaptureURL,
		Channel:     w.cfg.WebhookChannel,
		Caption:     caption,
		ChatID:      w.cfg.WebhookChatID,
		WARecipient: w.cfg.WebhookWARecipient,
	}

	go func() {
		data, err := json.Marshal(payload)
		if err != nil {
			log.Warn().Err(err).Msg("failed to marshal server webhook payload")
			return
		}

		req, err := http.NewRequest("POST", w.cfg.WebhookURL, bytes.NewReader(data))
		if err != nil {
			log.Warn().Err(err).Msg("failed to create server webhook request")
			return
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := w.httpClient.Do(req)
		if err != nil {
			log.Warn().Err(err).Str("url", w.cfg.WebhookURL).Msg("failed to send server monitoring webhook alert")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			log.Warn().Int("status", resp.StatusCode).Msg("server monitoring webhook returned error status")
		} else {
			log.Info().Int("status", resp.StatusCode).Msg("server monitoring webhook alert sent successfully")
		}
	}()
}

// Check evaluates the snapshot against thresholds and manages debounced alert triggers.
func (w *WebhookNotifier) Check(metrics *model.ServerMetrics) {
	if metrics == nil || !w.cfg.WebhookEnabled || w.cfg.WebhookURL == "" {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()

	cpuThreshold := w.cfg.AlertCpuThreshold
	if cpuThreshold <= 0 {
		cpuThreshold = 95.0
	}
	memThreshold := w.cfg.AlertMemThreshold
	if memThreshold <= 0 {
		memThreshold = 95.0
	}
	debounceTicks := w.cfg.AlertDebounceTicks
	if debounceTicks <= 0 {
		debounceTicks = 15
	}
	cooldown := w.cfg.AlertCooldown
	if cooldown <= 0 {
		cooldown = 15 * time.Minute
	}

	isCritical := metrics.Cpu.Usage >= cpuThreshold || metrics.Memory.Usage >= memThreshold

	// If critical load reached (>= 95%)
	if isCritical {
		w.critCount++
		w.healthyCount = 0

		// Require sustained high load for debounceTicks (e.g. 15s) before triggering alert
		if w.critCount >= debounceTicks {
			if !w.isAlerting || now.Sub(w.lastAlertTime) >= cooldown {
				w.isAlerting = true
				if w.alertedAt.IsZero() {
					w.alertedAt = now
				}
				w.lastAlertTime = now
				w.sendAlert(metrics, now, cpuThreshold, memThreshold)
			}
		}
		return
	}

	// If below critical threshold
	w.critCount = 0

	// Check recovery if currently alerting: require load to drop below 90% for at least 5 ticks
	if w.isAlerting {
		if metrics.Cpu.Usage < 90.0 && metrics.Memory.Usage < 90.0 {
			w.healthyCount++
			if w.healthyCount >= 5 {
				downtime := now.Sub(w.alertedAt)
				w.isAlerting = false
				w.alertedAt = time.Time{}
				w.healthyCount = 0
				w.sendRecovery(metrics, downtime, now)
			}
		} else {
			w.healthyCount = 0
		}
	}
}

func (w *WebhookNotifier) sendAlert(m *model.ServerMetrics, at time.Time, cpuThreshold, memThreshold float64) {
	hostname := m.ServerID
	if m.Host != nil && m.Host.Hostname != "" {
		hostname = m.Host.Hostname
	}

	timeStr := at.Format("02 Jan 2006, 15:04:05 MST")
	topProc := "-"
	if len(m.Processes) > 0 {
		topProc = fmt.Sprintf("%s (PID: %d, CPU: %.1f%%, RAM: %.1f%%)",
			m.Processes[0].Name, m.Processes[0].PID, m.Processes[0].CpuPercent, m.Processes[0].MemoryPercent)
	}

	triggerReason := ""
	if m.Cpu.Usage >= cpuThreshold && m.Memory.Usage >= memThreshold {
		triggerReason = fmt.Sprintf("CPU (%.1f%%) & RAM (%.1f%%) >= %.0f%%", m.Cpu.Usage, m.Memory.Usage, cpuThreshold)
	} else if m.Cpu.Usage >= cpuThreshold {
		triggerReason = fmt.Sprintf("CPU Usage (%.1f%%) >= %.0f%%", m.Cpu.Usage, cpuThreshold)
	} else if m.Memory.Usage >= memThreshold {
		triggerReason = fmt.Sprintf("RAM Usage (%.1f%%) >= %.0f%%", m.Memory.Usage, memThreshold)
	} else {
		triggerReason = fmt.Sprintf("Status %s", m.Status)
	}

	caption := fmt.Sprintf(
		"🚨 <b>SERVER MONITOR: CRITICAL ALERT</b>\n"+
			"----------------------------------------\n"+
			"🖥️ <b>Server:</b> %s (%s)\n"+
			"📊 <b>Status:</b> 🔴 CRITICAL\n"+
			"⚠️ <b>Pemicu:</b> %s\n"+
			"⚡ <b>CPU Usage:</b> %.1f%%\n"+
			"💾 <b>RAM Usage:</b> %.1f%%\n"+
			"📈 <b>Load (1m/5m/15m):</b> %.2f / %.2f / %.2f\n"+
			"🔥 <b>Top Process:</b> %s\n"+
			"⏰ <b>Waktu:</b> %s\n"+
			"----------------------------------------",
		m.ServerID, hostname, triggerReason, m.Cpu.Usage, m.Memory.Usage,
		m.Load.Load1, m.Load.Load5, m.Load.Load15,
		topProc, timeStr,
	)

	w.send(caption)
}

func (w *WebhookNotifier) sendRecovery(m *model.ServerMetrics, downtime time.Duration, at time.Time) {
	hostname := m.ServerID
	if m.Host != nil && m.Host.Hostname != "" {
		hostname = m.Host.Hostname
	}

	timeStr := at.Format("02 Jan 2006, 15:04:05 MST")
	downtimeStr := formatDuration(downtime)

	caption := fmt.Sprintf(
		"✅ <b>SERVER MONITOR: RECOVERED</b>\n"+
			"----------------------------------------\n"+
			"🖥️ <b>Server:</b> %s (%s)\n"+
			"📊 <b>Status:</b> 🟢 HEALTHY (Normal)\n"+
			"⚡ <b>CPU Usage:</b> %.1f%%\n"+
			"💾 <b>RAM Usage:</b> %.1f%%\n"+
			"⏱️ <b>Durasi Masalah:</b> %s\n"+
			"⏰ <b>Waktu Pulih:</b> %s\n"+
			"----------------------------------------",
		m.ServerID, hostname, m.Cpu.Usage, m.Memory.Usage,
		downtimeStr, timeStr,
	)

	w.send(caption)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d detik", int(d.Seconds()))
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if minutes < 60 {
		return fmt.Sprintf("%d menit %d detik", minutes, seconds)
	}
	hours := minutes / 60
	remMinutes := minutes % 60
	return fmt.Sprintf("%d jam %d menit", hours, remMinutes)
}
