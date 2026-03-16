.PHONY: build run

build:
	@go build -o logmx ./cmd/logmx

run:
	@go run ./cmd/logmx
