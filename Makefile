.PHONY: all build test clean docker-up docker-down docker-build dev-backend dev-frontend test-backend test-frontend

# Default target
all: test build

# Docker commands
docker-up:
	docker compose up --build

docker-down:
	docker compose down

# Testing
test: test-backend test-frontend

test-backend:
	cd backend && go test ./... -v

test-frontend:
	cd frontend && npm test -- --run

# Development (Run locally without Docker)
# Note: Requires MongoDB running on localhost:27017 and Ollama running
dev-backend:
	cd backend && go run main.go

dev-frontend:
	cd frontend && npm install && npm run dev

# Building (Locally)
build: build-backend build-frontend

build-backend:
	cd backend && go build -o bin/main main.go

build-frontend:
	cd frontend && npm install && npm run build

# Clean
clean:
	rm -rf backend/bin
	rm -rf backend/tmp
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	cd backend && go clean
