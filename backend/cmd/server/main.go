package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"server-monitor/backend/internal/collector"
	"server-monitor/backend/internal/config"
	"server-monitor/backend/internal/handler"
	"server-monitor/backend/internal/notify"
	"server-monitor/backend/internal/store"
	"server-monitor/backend/internal/websocket"
)

func main() {
	// Configure human-friendly console logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg := config.Load()
	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().
		Str("server_id", cfg.ServerID).
		Str("addr", cfg.HTTPAddr).
		Dur("metrics_interval", cfg.MetricsInterval).
		Int("history_size", cfg.HistorySize).
		Msg("starting server-monitor agent")

	// 1. Initialize In-Memory Metrics Store
	memStore := store.NewMemoryStore(cfg.HistorySize)

	// 2. Initialize WebSocket Hub
	// The initial callback gives every newly connected dashboard an immediate
	// full snapshot so charts render instantly instead of waiting for the next
	// collector tick.
	wsHub := websocket.NewHub(func() any {
		latest, ok := memStore.GetLatest()
		if !ok {
			return nil
		}
		return latest
	})
	go wsHub.Run()

	// 3. Initialize & Start Metrics Collector Engine
	var alertNotifier notify.Notifier = notify.NopNotifier{}
	if cfg.WebhookEnabled {
		alertNotifier = notify.NewWebhookNotifier(cfg)
		log.Info().
			Str("webhook_url", cfg.WebhookURL).
			Str("channel", cfg.WebhookChannel).
			Str("capture_url", cfg.CaptureURL).
			Msg("monitoring alert webhook notifier enabled")
	}

	collectorCtx, cancelCollector := context.WithCancel(context.Background())
	engine := collector.NewEngine(cfg.ServerID, cfg.MetricsInterval, memStore, wsHub, alertNotifier)
	go engine.Start(collectorCtx)

	// 4. Initialize HTTP & WS Router
	router := handler.NewRouter(cfg, memStore, wsHub)

	srv := &http.Server{
		Addr:        cfg.HTTPAddr,
		Handler:     router,
		ReadTimeout: 15 * time.Second,
		// WriteTimeout intentionally omitted: it applies as an absolute deadline
		// to hijacked WebSocket connections, cutting every dashboard stream after
		// the timeout expires. Per-write deadlines are already enforced in
		// writePump via SetWriteDeadline.
		IdleTimeout: 60 * time.Second,
	}

	// 5. Start HTTP Server in background goroutine
	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("HTTP & WebSocket server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server encountered fatal error")
		}
	}()

	// 6. Wait for Termination Signals for Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server gracefully...")

	// Cancel collector loop
	cancelCollector()

	// Stop websocket hub
	wsHub.Stop()

	// Shutdown HTTP server with timeout
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server forced to shutdown")
	}

	log.Info().Msg("server-monitor exited cleanly")
}
