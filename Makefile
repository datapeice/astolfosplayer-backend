.PHONY: help proto build test clean docker-up docker-down docker-rebuild run-auth run-file run-sync install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  make proto          - Generate Go code from .proto files"
	@echo "  make build          - Build all services"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-up      - Start all services with Docker Compose"
	@echo "  make docker-down    - Stop all services"
	@echo "  make docker-rebuild - Rebuild and restart all services"
	@echo "  make run-auth       - Run Auth service locally"
	@echo "  make run-file       - Run File service locally"
	@echo "  make run-sync       - Run Sync service locally"
	@echo "  make install-tools  - Install required tools (protoc plugins)"

# Install required tools
install-tools:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from proto files
proto:
	@echo "Generating Go code from .proto files..."
	rm -rf protos/gen/go
	mkdir -p protos/gen/go
	export PATH=$$PATH:$$(go env GOPATH)/bin && \
	protoc -I protos/proto \
		--go_out=protos/gen/go \
		--go_opt=paths=source_relative \
		--go-grpc_out=protos/gen/go \
		--go-grpc_opt=paths=source_relative \
		protos/proto/auth/auth.proto \
		protos/proto/file/file.proto \
		protos/proto/sync/sync.proto
	@echo "Proto generation complete!"

# Build all services
build:
	@echo "Building all services..."
	mkdir -p bin
	CGO_ENABLED=1 go build -o bin/auth-service cmd/auth/main.go
	CGO_ENABLED=1 go build -o bin/file-service cmd/file/main.go
	CGO_ENABLED=1 go build -o bin/sync-service cmd/sync/main.go
	@echo "Build complete!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf protos/gen/go/
	rm -f *.db
	@echo "Clean complete!"

# Docker Compose commands
docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d
	@echo "Services started!"
	@echo "Auth Service: localhost:50051"
	@echo "File Service: localhost:50052"
	@echo "Sync Service: localhost:50053"
	@echo "MinIO Console: http://localhost:9001 (admin: minioadmin/minioadmin)"

docker-down:
	@echo "Stopping services..."
	docker-compose down

docker-rebuild:
	@echo "Rebuilding and restarting services..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
	@echo "Services restarted!"

# Run services locally (development)
run-auth:
	@echo "Running Auth Service..."
	export DATABASE_URL=auth.db && \
	export SECRET_KEY=dev-secret-key && \
	export SECURITY_KEY=dev-security-key && \
	export PORT=50051 && \
	go run cmd/auth/main.go

run-file:
	@echo "Running File Service..."
	export DATABASE_URL=metadata.db && \
	export S3_ENDPOINT=localhost:9000 && \
	export S3_ACCESS_KEY=minioadmin && \
	export S3_SECRET_KEY=minioadmin && \
	export S3_BUCKET=music && \
	export S3_USE_SSL=false && \
	export PORT=50052 && \
	go run cmd/file/main.go

run-sync:
	@echo "Running Sync Service..."
	export DATABASE_URL=metadata.db && \
	export PORT=50053 && \
	go run cmd/sync/main.go

# Development workflow
dev: proto build

# Full cycle: clean, proto, build
all: clean proto build
