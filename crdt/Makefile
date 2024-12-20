.PHONY: all build run

all: build run

build:
	@go build

run:
	@./crdt --port 8080 localhost:8081 localhost:8082&
	@./crdt --port 8081 localhost:8080 localhost:8082&
	@./crdt --port 8082 localhost:8080 localhost:8081&