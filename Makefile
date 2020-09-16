.PHONY: build lint test vendor run

default: build

build:
	@go build -mod=vendor -o build/cached -ldflags "-s -w" cmd/main.go

lint:
	@golangci-lint run

test:
	@go test ./...

vendor:
	@go mod tidy && go mod vendor

run:
	@go run cmd/main.go
