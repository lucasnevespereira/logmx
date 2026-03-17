VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build run clean

build:
	@go build -ldflags "$(LDFLAGS)" -o bin/logmx ./cmd/logmx

run:
	@go run ./cmd/logmx

clean:
	@rm -rf bin
