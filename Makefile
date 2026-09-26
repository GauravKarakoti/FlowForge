.PHONY: up-infra run-server run-worker test lint

up-infra:
# 	TODO: Start Postgres and Redis via docker-compose
	docker-compose up -d

run-server:
# 	TODO: Build and run the control plane server
	go run cmd/server/main.go

run-worker:
# 	TODO: Build and run a local worker node
	go run cmd/worker/main.go

test:
# 	TODO: Implement test runner
	go test ./...

lint:
# 	TODO: Add golangci-lint execution
	go vet ./...