os ?= linux
arch ?= amd64

.DEFAULT_GOAL := build

.PHONY: mod fmt vet build clean

mod:
	go mod tidy

fmt: mod
	go fmt ./...

vet: fmt
	go vet ./...

build: fmt
	@if [ -z "$(ver)" ]; then \
		echo "Must specify version number"; \
		exit 1; \
	fi

	GOOS=$(os) GOARCH=$(arch) go build -o nibl ./cmd/cli/main.go

	zip nibl_$(ver)_$(os)_$(arch).zip nibl
	rm -f ./nibl

local: fmt
	go build -o nibl ./cmd/cli/main.go

clean:
	go clean ./...
