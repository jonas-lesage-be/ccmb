export CGO_ENABLED=0

.PHONY: all
all: tidy fmt lint cover build

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	golangci-lint fmt

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test: export CGO_ENABLED=1
test:
	go test -v -race -timeout 30s -coverprofile=coverage.out ./...

.PHONY: cover
cover: test
	go tool cover -html=coverage.out

.PHONY: build
build:
	go build -ldflags="-s -w" .
