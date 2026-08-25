.PHONY: lint run version

lint:
	golangci-lint run ./...

run:
	clear && go run ./cmd/p-manager/main.go

version:
	go run ./cmd/vault/main.go -v
