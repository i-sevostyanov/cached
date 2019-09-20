.PHONY: build lint test vendor

default: build

build:
	go build -mod=vendor -o build/probe -ldflags "-s -w" cmd/main.go

lint:
	golangci-lint run

test:
	go test ./...

vendor:
	go mod tidy && go mod vendor
