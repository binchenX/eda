.PHONY: up down run send-event clean

# Start infrastructure (Kafka, Zookeeper, Schema Registry, etc.)
up:
	@echo "Starting infrastructure..."
	@./scripts/start.sh

# Stop all services
down:
	@echo "Stopping all services..."
	@docker-compose down

# Run all microservices
run:
	@echo "Starting microservices..."
	@cd OrderService && go run *.go & \
	cd InventorService && go run *.go & \
	cd ShppingService && go run *.go & \
	cd PaymentService && go run *.go & \
	echo "All services started. Use 'ps' to check running services."

# Send a test order event
send-event:
	@echo "Sending test order..."
	@curl -X POST \
		-H "Content-Type: application/json" \
		-d '{"userId":"test-user","items":["item1","item2"]}' \
		http://localhost:8080/order

# Kill all running Go services
clean:
	@echo "Cleaning up Go services..."
	@pkill -f "OrderService|InventorService|ShppingService|PaymentService" || true

# Helper target to show service status
status:
	@echo "Docker services:"
	@docker-compose ps
	@echo "\nGo services:"
	@ps aux | grep -E "OrderService|InventorService|ShppingService|PaymentService" | grep -v grep || echo "No Go services running"

# Show help
help:
	@echo "Available targets:"
	@echo "  up          - Start infrastructure (Kafka, Zookeeper, etc.)"
	@echo "  down        - Stop all Docker services"
	@echo "  run         - Start all microservices"
	@echo "  send-event  - Send a test order"
	@echo "  clean       - Kill all running Go services"
	@echo "  status      - Show status of all services"
	@echo "  help        - Show this help"

# Default target
.DEFAULT_GOAL := help 
