.PHONY: lint run

lint:
	golangci-lint run ./...

run:
	clear && go run ./cmd/vault/main.go