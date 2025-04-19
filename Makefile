.PHONY: build test integration-test test-coverage migrate run docker-build docker-run

build:
	go build -o bin/pvz-api ./cmd/api/main.go

test:
	go test -v ./...

integration-test:
	go test -v ./test/integration/...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total

migrate:
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/pvz_service?sslmode=disable" up

run:
	go run ./cmd/api/main.go

docker-build:
	docker build -t pvz-service .

docker-run:
	docker-compose up -d


