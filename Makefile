.DEFAULT_GOAL := build

.PHONY: mod fmt vet build clean

mod:
	go mod tidy

fmt: mod
	go fmt ./...

vet: fmt
	go vet ./...

build: fmt
	go build -o nibl ./cmd/cli/main.go

clean:
	go clean ./...
