.PHONY: lint run

lint:
	golangci-lint run ./...

run:
	go run ./cmd/vault/main.go