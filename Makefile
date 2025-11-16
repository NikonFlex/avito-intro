.PHONY: build run

BINARY_NAME=server

build:
	go build -o bin/$(BINARY_NAME) cmd/pr-reviewer/main.go

run: build
	./bin/$(BINARY_NAME)
