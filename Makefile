.PHONY: build run test clean docker-up docker-down lint

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-up:
	docker compose up -d

docker-down:
	docker compose down

lint:
	golangci-lint run
