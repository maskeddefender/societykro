.PHONY: help setup dev dev-db dev-services dev-mobile dev-web docker-up docker-down migrate migrate-down seed test test-go test-ts lint clean generate proto sqlc

# ==========================================
# HELP
# ==========================================
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ==========================================
# SETUP
# ==========================================
setup: ## First-time setup: install all dependencies
	@echo "==> Installing Go dependencies..."
	cd services/auth-service && go mod download
	cd services/complaint-service && go mod download
	cd services/visitor-service && go mod download
	cd services/payment-service && go mod download
	cd services/notice-service && go mod download
	cd services/vendor-service && go mod download
	cd services/chatbot-service && go mod download
	cd services/notification-service && go mod download
	cd services/message-router && go mod download
	cd services/voice-service && go mod download
	@echo "==> Installing mobile dependencies..."
	cd apps/mobile && pnpm install
	@echo "==> Installing web-admin dependencies..."
	cd apps/web-admin && pnpm install
	@echo "==> Generating RSA keys for JWT..."
	mkdir -p keys
	openssl genrsa -out keys/private.pem 2048
	openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "==> Copying .env..."
	cp -n .env.example .env || true
	@echo "==> Setup complete!"

# ==========================================
# DOCKER (Local databases)
# ==========================================
docker-up: ## Start all local databases (PG, Redis, NATS, MinIO)
	docker compose up -d postgres redis nats minio minio-init
	@echo "==> Waiting for services..."
	@sleep 3
	@echo "==> PostgreSQL: localhost:5432"
	@echo "==> Redis:      localhost:6379"
	@echo "==> NATS:       localhost:4222 (monitor: localhost:8222)"
	@echo "==> MinIO:      localhost:9000 (console: localhost:9001)"

docker-down: ## Stop all local databases
	docker compose down

docker-clean: ## Stop and remove all data volumes
	docker compose down -v

docker-tools: ## Start optional tools (pgAdmin)
	docker compose --profile tools up -d pgadmin
	@echo "==> pgAdmin: http://localhost:5050 (admin@societykro.in / admin)"

# ==========================================
# DATABASE
# ==========================================
migrate: ## Run all database migrations
	@echo "==> Running migrations..."
	@for f in db/migrations/*.sql; do \
		echo "  Applying $$f..."; \
		PGPASSWORD=societykro_dev psql -h localhost -U societykro -d societykro -f $$f; \
	done
	@echo "==> Migrations complete!"

migrate-fresh: ## Drop all tables and re-run migrations (DESTRUCTIVE)
	@echo "==> Dropping all tables..."
	PGPASSWORD=societykro_dev psql -h localhost -U societykro -d societykro -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) migrate

seed: ## Seed database with test data
	@echo "==> Seeding database..."
	@for f in db/seeds/*.sql; do \
		echo "  Seeding $$f..."; \
		PGPASSWORD=societykro_dev psql -h localhost -U societykro -d societykro -f $$f; \
	done
	@echo "==> Seed complete!"

# ==========================================
# DEVELOPMENT
# ==========================================
dev: docker-up migrate ## Start everything (databases + services + apps)
	@echo "==> Starting all services..."
	@$(MAKE) dev-services &
	@$(MAKE) dev-mobile &
	@$(MAKE) dev-web &
	@wait

dev-services: ## Start all Go backend services
	@echo "==> Starting Go services..."
	cd services/auth-service && go run cmd/server/main.go &
	cd services/complaint-service && go run cmd/server/main.go &
	cd services/visitor-service && go run cmd/server/main.go &
	cd services/payment-service && go run cmd/server/main.go &
	cd services/notice-service && go run cmd/server/main.go &
	cd services/notification-service && go run cmd/server/main.go &
	cd services/message-router && go run cmd/server/main.go &
	cd services/voice-service && go run cmd/server/main.go &

dev-mobile: ## Start React Native mobile app
	cd apps/mobile && pnpm start

dev-web: ## Start Next.js web admin
	cd apps/web-admin && pnpm dev

# ==========================================
# CODE GENERATION
# ==========================================
generate: proto sqlc ## Generate all code (proto + sqlc)

proto: ## Generate Go code from protobuf definitions
	@echo "==> Generating protobuf Go code..."
	@find packages/proto -name "*.proto" -exec protoc --go_out=. --go-grpc_out=. {} \;
	@echo "==> Proto generation complete!"

sqlc: ## Generate Go code from SQL queries
	@echo "==> Generating sqlc Go code..."
	cd db && sqlc generate
	@echo "==> sqlc generation complete!"

# ==========================================
# TESTING
# ==========================================
test: test-go test-ts ## Run all tests

test-go: ## Run Go tests
	@echo "==> Running Go tests..."
	@for svc in services/*/; do \
		echo "  Testing $$svc..."; \
		cd $$svc && go test ./... -v -count=1 && cd ../..; \
	done

test-ts: ## Run TypeScript tests
	cd apps/mobile && pnpm test
	cd apps/web-admin && pnpm test

# ==========================================
# LINTING
# ==========================================
lint: ## Run all linters
	@echo "==> Linting Go..."
	@for svc in services/*/; do \
		cd $$svc && golangci-lint run ./... && cd ../..; \
	done
	@echo "==> Linting TypeScript..."
	cd apps/mobile && pnpm lint
	cd apps/web-admin && pnpm lint

# ==========================================
# BUILD
# ==========================================
build-services: ## Build all Go service binaries
	@for svc in services/*/; do \
		name=$$(basename $$svc); \
		echo "  Building $$name..."; \
		cd $$svc && CGO_ENABLED=0 go build -o ../../bin/$$name cmd/server/main.go && cd ../..; \
	done
	@echo "==> All services built in ./bin/"

# ==========================================
# CLEANUP
# ==========================================
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf apps/mobile/.expo
	rm -rf apps/web-admin/.next
	@echo "==> Cleaned!"
