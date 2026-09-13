export CGO_ENABLED=0

.PHONY: all tidy fmt lint test cover build

all: tidy fmt lint cover build

tidy:
	go mod tidy

fmt:
	golangci-lint fmt

lint:
	golangci-lint run

test: export CGO_ENABLED=1
test:
	go test -v -race -timeout 30s -coverprofile=coverage.out ./...

cover: test
	go tool cover -html=coverage.out

build:
	go build -ldflags="-s -w" .
