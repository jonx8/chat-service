.PHONY: run build test docker-up docker-down fmt vet lint clean

all: build

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/app ./cmd/server

test:
	go test ./...

# Docker commands
docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down


# Code quality
fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...
	
clean:
	rm -rf bin/