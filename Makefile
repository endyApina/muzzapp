PROTO_SRC=proto/explore-service.proto
PROTO_OUT=proto/gen
MOCKERY=github.com/vektra/mockery/v2@v2.52.2

.PHONY: all help build start-services stop-services restart clean test generate-protos generate-mocks

help:
	@echo "Available commands:"
	@echo "  make build              Build Go binary and Docker image"
	@echo "  make start-services     Start all services (DB, Redis, App)"
	@echo "  make stop-services      Stop all services"
	@echo "  make restart            Restart services"
	@echo "  make clean              Remove built binary, containers, images, and volumes"
	@echo "  make test               Run Go tests"
	@echo "  make generate-protos    Generate Go code from protobuf definitions"
	@echo "  make generate-mocks     Generate Go mocks using mockery"

# Build Go binary and Docker image
build:
	@echo "Building Go binary and Docker image..."
	go mod tidy
	go build -o muzzapp ./cmd
	docker-compose build

# Start services
start-services: generate-protos generate-mocks test build
	@echo "Starting all services..."
	docker-compose up

# Stop services
stop-services:
	@echo "Stopping all services..."
	docker-compose down

# Restart services
restart: stop-services start-services

# Clean everything
clean:
	@echo "Cleaning up..."
	rm -f muzzapp
	docker-compose down --rmi local -v

# Run Go tests
test: generate-mocks
	go test ./... -v

# Generate protobuf Go files
generate-protos:
	@echo "Generating protobuf Go files..."
	protoc --go_out=$(PROTO_OUT) --go-grpc_out=$(PROTO_OUT) $(PROTO_SRC)

# Generate mocks with mockery
generate-mocks:
	@echo "Generating mocks..."
	go install $(MOCKERY)
	go run $(MOCKERY)
