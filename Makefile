check:
	go fmt ./...
	go vet ./...
	go test ./...

lint:
	golangci-lint run

kernel-selftest:
	go run ./cmd/hg check --selftest
