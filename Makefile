.PHONY: build clean test test-race test-cubical coverage lint check kernel-selftest help

build:
	go build -o bin/hg ./cmd/hg

clean:
	rm -rf bin/ coverage.out coverage.html

test:
	go test ./...

test-race:
	go test -race ./...

test-cubical:
	go test -tags cubical ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

check:
	go fmt ./...
	go vet ./...
	go test ./...

kernel-selftest:
	go run ./cmd/hg check --selftest

help:
	@echo "Available targets:"
	@echo "  build          - Build the hg binary"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-cubical   - Run cubical feature tests"
	@echo "  coverage       - Generate coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  check          - Run fmt, vet, and tests"
	@echo "  kernel-selftest - Run kernel self-tests"
