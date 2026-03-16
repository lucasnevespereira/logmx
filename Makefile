.PHONY: build run

build:
	@go build -o bin/logmx ./cmd/logmx

run:
	@go run ./cmd/logmx
