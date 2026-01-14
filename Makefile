.PHONY: run build test fmt vet lint clean

all: build

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/app ./cmd/server

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...
	
clean:
	rm -rf bin/