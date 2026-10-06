package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents runtime backend configuration.
type Config struct {
	HTTPAddr        string
	ServerID        string
	APIToken        string
	CorsOrigins     []string
	MetricsInterval time.Duration
	HistorySize     int
	LogLevel        string

	// Webhook / Alerting
	WebhookEnabled     bool
	WebhookURL         string
	WebhookChannel     string
	WebhookChatID      string
	WebhookWARecipient string
	CaptureURL         string

	// Thresholds for alert
	AlertCpuThreshold  float64
	AlertMemThreshold  float64
	AlertDebounceTicks int
	AlertCooldown      time.Duration
}

// Load populates configuration from CLI flags and environment variables with sensible defaults.
func Load() *Config {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "server-local"
	}

	// Environment variable fallbacks
	envServerID := getEnv("SERVER_ID", hostname)
	envHTTPAddr := getEnv("HTTP_ADDR", ":9191")
	envAPIToken := getEnv("API_TOKEN", "")
	envOrigins := getEnv("CORS_ORIGINS", "*")
	envInterval := getEnv("METRICS_INTERVAL", "1s")
	envHistorySize := getEnv("HISTORY_SIZE", "60")
	envLogLevel := getEnv("LOG_LEVEL", "info")

	// CLI flags (flags override env vars if specified)
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagHTTPAddr := flagSet.String("addr", envHTTPAddr, "HTTP and WebSocket listen address (e.g. :9191 or 0.0.0.0:9191)")
	flagServerID := flagSet.String("server-id", envServerID, "Unique server identifier")
	flagAPIToken := flagSet.String("token", envAPIToken, "Optional API Bearer authentication token")
	flagOrigins := flagSet.String("cors", envOrigins, "Comma-separated allowed CORS origins")
	flagInterval := flagSet.String("interval", envInterval, "Metrics polling interval (e.g. 1s, 500ms)")
	flagHistorySize := flagSet.Int("history", 0, "Number of rolling historical samples in memory (default 60)")
	flagLogLevel := flagSet.String("log-level", envLogLevel, "Log verbosity level (debug, info, warn, error)")

	_ = flagSet.Parse(os.Args[1:])

	httpAddr := *flagHTTPAddr
	serverID := *flagServerID
	apiToken := *flagAPIToken
	corsOrigins := parseCommaList(*flagOrigins)

	interval, err := time.ParseDuration(*flagInterval)
	if err != nil || interval < 250*time.Millisecond {
		interval = 1 * time.Second
	}

	historySize := *flagHistorySize
	if historySize <= 0 {
		hs, err := strconv.Atoi(envHistorySize)
		if err != nil || hs < 10 {
			historySize = 60
		} else {
			historySize = hs
		}
	}

	logLevel := *flagLogLevel

	cooldownMinutes := getInt("ALERT_COOLDOWN_MINUTES", 15)
	if cooldownMinutes <= 0 {
		cooldownMinutes = 15
	}

	return &Config{
		HTTPAddr:        httpAddr,
		ServerID:        serverID,
		APIToken:        apiToken,
		CorsOrigins:     corsOrigins,
		MetricsInterval: interval,
		HistorySize:     historySize,
		LogLevel:        logLevel,

		WebhookEnabled:     getBool("WEBHOOK_ENABLED", false),
		WebhookURL:         getEnv("WEBHOOK_URL", ""),
		WebhookChannel:     getEnv("WEBHOOK_CHANNEL", "both"),
		WebhookChatID:      getEnv("WEBHOOK_CHAT_ID", ""),
		WebhookWARecipient: getEnv("WEBHOOK_WA_RECIPIENT", ""),
		CaptureURL:         getEnv("CAPTURE_URL", "http://localhost:8080/"),

		AlertCpuThreshold:  getFloat("ALERT_CPU_THRESHOLD", 95.0),
		AlertMemThreshold:  getFloat("ALERT_MEM_THRESHOLD", 95.0),
		AlertDebounceTicks: getInt("ALERT_DEBOUNCE_TICKS", 15),
		AlertCooldown:      time.Duration(cooldownMinutes) * time.Minute,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return val
}

func getFloat(key string, fallback float64) float64 {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil || val <= 0 {
		return fallback
	}
	return val
}

func getInt(key string, fallback int) int {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return fallback
	}
	return val
}

func parseCommaList(raw string) []string {
	if strings.TrimSpace(raw) == "" || raw == "*" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}
	if len(res) == 0 {
		return []string{"*"}
	}
	return res
}
