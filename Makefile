.PHONY: help build-backend build-frontend build-all test-backend run-backend run-frontend clean docker-up docker-down docker-restart docker-logs docker-rebuild

help:
	@echo "Distributed Server Monitoring Dashboard - Makefile"
	@echo "----------------------------------------------------"
	@echo "make build-backend   - Build Go backend agent binary"
	@echo "make test-backend    - Run Go backend unit tests"
	@echo "make build-frontend  - Build static Vue 3 production dist/"
	@echo "make build-all       - Build both backend binary and frontend dist/"
	@echo "make run-backend     - Run Go backend locally"
	@echo "make run-frontend    - Run Vite frontend dev server"
	@echo "make clean           - Remove build artifacts"
	@echo ""
	@echo "Docker (central dashboard server):"
	@echo "make docker-up       - Build image and start the dashboard container"
	@echo "make docker-down     - Stop and remove the dashboard container"
	@echo "make docker-restart  - Restart to pick up .env changes (no rebuild)"
	@echo "make docker-rebuild  - Rebuild from scratch after changing source code"
	@echo "make docker-logs     - Follow container logs"

build-backend:
	@echo "==> Building Go backend..."
	cd backend && go build -ldflags="-s -w" -o bin/server-monitor ./cmd/server

test-backend:
	@echo "==> Running Go backend tests..."
	cd backend && go test ./... -v

build-frontend:
	@echo "==> Building Vue 3 static distribution..."
	cd frontend && npm run build

build-all: test-backend build-backend build-frontend
	@echo "==> All builds completed successfully!"

run-backend:
	@echo "==> Starting Go monitoring agent on :8080..."
	cd backend && go run ./cmd/server/main.go

run-frontend:
	@echo "==> Starting Vite development server..."
	cd frontend && npm run dev

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf backend/bin frontend/dist

# --- Docker (central dashboard server) -------------------------------------

docker-up:
	@echo "==> Starting dashboard container..."
	docker compose up -d --build
	@echo "==> Up. Check with: make docker-logs"

docker-down:
	@echo "==> Stopping dashboard container..."
	docker compose down

# Agent addresses and the published port are read at container start, so an
# .env change only needs a restart, never an image rebuild.
docker-restart:
	@echo "==> Restarting to apply .env changes..."
	docker compose up -d

docker-rebuild:
	@echo "==> Rebuilding image without cache..."
	docker compose build --no-cache
	docker compose up -d

docker-logs:
	docker compose logs -f
