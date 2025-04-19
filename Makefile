.PHONY: build generate-proto test integration-test migrate run docker-build docker-run

build:
	go build -o bin/pvz-api ./cmd/api/main.go

generate-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/pvz.proto

test:
	go test -v ./...

integration-test:
	go test -v ./test/integration/...

migrate:
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/pvz_service?sslmode=disable" up

run:
	go run ./cmd/api/main.go

docker-build:
	docker build -t pvz-service .

docker-run:
	docker-compose up -d

