.PHONY: dev backend frontend build clean loadtest

# Development
dev: backend-dev frontend-dev

backend-dev:
	cd backend && go run cmd/server/main.go

frontend-dev:
	cd frontend && npm run dev

# Build
build-frontend:
	cd frontend && npm run build

build-backend: build-frontend
	cd backend && CGO_ENABLED=1 go build -o ../iot-admin-server cmd/server/main.go

build: build-backend
	@echo "Build complete: ./iot-admin-server"

# Run
run: build
	./iot-admin-server

# Clients
client-go-pub:
	cd clients/go/publisher && go run main.go

client-go-sub:
	cd clients/go/subscriber && go run main.go

client-py-pub:
	cd clients/python && python3 publisher.py

client-py-sub:
	cd clients/python && python3 subscriber.py

# Load test
loadtest:
	cd loadtest && go run main.go

loadtest-100:
	NUM_DEVICES=100 MSG_INTERVAL_MS=1000 DURATION_SECS=60 go run ./loadtest/main.go

# Clean
clean:
	rm -f iot-admin-server
	rm -rf backend/data
	rm -rf frontend/dist
	find . -name "*.db" -delete
