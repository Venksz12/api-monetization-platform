PROTOC=protoc
GO=go

.PHONY: proto tidy build test test-race run-gateway run-wallet run-metering run-billing

proto:
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative proto/wallet.proto

tidy:
	$(GO) mod tidy

build:
	$(GO) build ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

run-gateway:
	$(GO) run ./cmd/gateway

run-wallet:
	$(GO) run ./cmd/wallet

run-metering:
	$(GO) run ./cmd/metering

run-billing:
	$(GO) run ./cmd/billing
