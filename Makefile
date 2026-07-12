.PHONY: help build test run run-gateway run-backends run-infra run-all docker-build docker-up docker-down lint proto coverage bench-basic clean

help:
	@echo "API Gateway Makefile Commands"
	@echo "=============================="
	@echo "build              - Build the gateway binary"
	@echo "test               - Run all unit tests"
	@echo "coverage           - Run tests with coverage report"
	@echo "lint               - Run golangci-lint"
	@echo "proto              - Generate protobuf files"
	@echo "run-gateway        - Run gateway locally (requires Redis running)"
	@echo "run-backends       - Start backend services with Docker Compose"
	@echo "run-infra          - Start Redis, Prometheus, Grafana with Docker Compose"
	@echo "run-all            - Start all services with Docker Compose"
	@echo "docker-build       - Build gateway Docker image"
	@echo "docker-up          - Start all services with Docker Compose"
	@echo "docker-down        - Stop and remove all Docker Compose services"
	@echo "bench-basic        - Run k6 basic load test"
	@echo "clean              - Clean build artifacts"

build:
	go build -o bin/gateway ./cmd/gateway

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	golangci-lint run ./...

proto:
	protoc --go_out=. --go-grpc_out=. backends/service-c/proto/echo.proto

run-gateway: build
	./bin/gateway

run-backends:
	docker compose up service-a service-b service-c service-d

run-infra:
	docker compose up redis prometheus grafana

run-all:
	docker compose up

docker-build:
	docker build -t api-gateway:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

bench-basic:
	k6 run benchmarks/k6/basic_load.js

clean:
	rm -rf bin/ coverage.out coverage.html
