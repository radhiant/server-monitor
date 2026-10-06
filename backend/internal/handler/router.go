package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"server-monitor/backend/internal/config"
	"server-monitor/backend/internal/model"
	"server-monitor/backend/internal/store"
	"server-monitor/backend/internal/websocket"
)

// NewRouter constructs and configures the HTTP & WebSocket router.
func NewRouter(cfg *config.Config, s store.MetricsStore, hub *websocket.Hub) http.Handler {
	r := chi.NewRouter()

	// Base middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// CORS Setup
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Handlers
	healthHandler := NewHealthHandler(cfg)
	serverHandler := NewServerHandler(s)
	metricsHandler := NewMetricsHandler(s)
	processHandler := NewProcessHandler(s)
	promHandler := NewPrometheusHandler(s, cfg)

	// Auth middleware wrapper
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.APIToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check Authorization Header
			authHeader := r.Header.Get("Authorization")
			token := ""
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else if queryToken := r.URL.Query().Get("token"); queryToken != "" {
				token = queryToken
			}

			if token != cfg.APIToken {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(model.ErrorResponse{
					Error: model.ErrorDetail{
						Code:    "UNAUTHORIZED",
						Message: "Invalid or missing API token",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Public Health & Prometheus Endpoints
	r.Get("/api/v1/health", healthHandler.ServeHTTP)
	r.Get("/metrics", promHandler.ServeHTTP)

	// Authenticated API group
	r.Group(func(api chi.Router) {
		api.Use(authMiddleware)

		api.Get("/api/v1/server", serverHandler.GetInfo)
		api.Get("/api/v1/status", serverHandler.GetStatus)
		api.Get("/api/v1/metrics", metricsHandler.GetLatest)
		api.Get("/api/v1/history", metricsHandler.GetHistory)
		api.Get("/api/v1/filesystems", metricsHandler.GetFilesystems)
		api.Get("/api/v1/network/interfaces", metricsHandler.GetInterfaces)
		api.Get("/api/v1/processes", processHandler.GetProcesses)

		// WebSocket Endpoint
		api.Get("/ws/v1", func(w http.ResponseWriter, r *http.Request) {
			websocket.ServeWs(hub, w, r)
		})
	})

	return r
}
