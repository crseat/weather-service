SHELL := /bin/bash

APP := weatherd

.PHONY: run test lint build docker-build

run:
	go run ./cmd/weatherd

test:
	go test ./... -race

lint:
	golangci-lint run --config ./.golangci.yml

build:
	go build -ldflags="-s -w -X weather-service/internal/version.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev) -X weather-service/internal/version.Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo none) -X weather-service/internal/version.BuiltAt=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/$(APP) ./cmd/weatherd

docker-build:
	docker build -t weather-service:local .

docker-run:
	docker run --rm -p 8080:8080 weather-service:local
