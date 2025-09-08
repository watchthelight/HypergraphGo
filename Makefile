SHELL := /bin/sh
.PHONY: tidy lint test test-race bench examples ci

tidy:
	go mod tidy

lint:
	golangci-lint run

test:
	go test ./... -count=1

test-race:
	CGO_ENABLED=1 go test ./... -race -count=1

bench:
	go test -bench=. -benchmem ./...

examples:
	go run ./examples/basic

ci: tidy lint test
