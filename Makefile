# ============================================
# Makefile for elotus_test
# ============================================

.PHONY: help build up down logs restart clean test dev

# Default target
help:
	@echo "╔════════════════════════════════════════════════════════════╗"
	@echo "║              elotus_test - Docker Commands                 ║"
	@echo "╠════════════════════════════════════════════════════════════╣"
	@echo "║  make build     - Build Docker images                      ║"
	@echo "║  make up        - Start all services                       ║"
	@echo "║  make down      - Stop all services                        ║"
	@echo "║  make restart   - Restart all services                     ║"
	@echo "║  make logs      - View logs (follow mode)                  ║"
	@echo "║  make logs-app  - View app logs only                       ║"
	@echo "║  make clean     - Remove all containers and volumes        ║"
	@echo "║  make test      - Run tests                                ║"
	@echo "║  make dev       - Start in development mode                ║"
	@echo "║  make shell     - Open shell in app container              ║"
	@echo "║  make psql      - Connect to PostgreSQL                    ║"
	@echo "║  make redis-cli - Connect to Redis                         ║"
	@echo "╚════════════════════════════════════════════════════════════╝"

# Build Docker images
build:
	@echo "🔨 Building Docker images..."
	docker-compose build

# Start all services
up:
	@echo "🚀 Starting all services..."
	docker-compose up -d
	@echo "✅ Services started!"
	@echo "   App:      http://localhost:8080"
	@echo "   Postgres: localhost:5432"
	@echo "   Redis:    localhost:6379"

# Start with logs
up-logs:
	@echo "🚀 Starting all services with logs..."
	docker-compose up

# Stop all services
down:
	@echo "🛑 Stopping all services..."
	docker-compose down

# Restart all services
restart: down up

# View all logs
logs:
	docker-compose logs -f

# View app logs only
logs-app:
	docker-compose logs -f app

# Remove everything (containers, volumes, images)
clean:
	@echo "🧹 Cleaning up everything..."
	docker-compose down -v --rmi local
	@echo "✅ Cleanup complete!"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./server/tests/... -v

# Development mode (rebuild and start)
dev:
	@echo "🔧 Starting development mode..."
	docker-compose up --build

# Open shell in app container
shell:
	docker-compose exec app sh

# Connect to PostgreSQL
psql:
	docker-compose exec postgres psql -U elotus -d elotus_test

# Connect to Redis CLI
redis-cli:
	docker-compose exec redis redis-cli -a redis_secret_password

# Check status of all services
status:
	@echo "📊 Service Status:"
	docker-compose ps

# View resource usage
stats:
	docker stats elotus-app elotus-postgres elotus-redis

