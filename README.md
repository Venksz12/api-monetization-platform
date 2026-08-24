# API Monetization & Usage Platform

Production-oriented distributed systems project in Go.

## Stack

- Go
- HTTP reverse proxy
- gRPC + Protobuf
- NATS JetStream
- Couchbase
- Redis
- Razorpay
- OpenTelemetry
- Prometheus
- Docker
- Kubernetes

## Capabilities

### Gateway
- SHA-256 API-key hashing
- Couchbase API-key persistence
- Redis distributed token-bucket rate limiting
- Burst support and HTTP 429
- X-Request-ID and X-Correlation-ID propagation
- Reverse proxy
- Usage event publication

### Usage
- NATS JetStream durable pull consumers
- explicit ACK after persistence
- exponential NAK backoff
- DLQ after max delivery attempts
- monthly owner usage rollups

### Wallet
- gRPC service
- integer minor-unit money
- Couchbase multi-document transactions
- immutable ledger
- persistent idempotency
- insufficient-balance protection

### Billing
- tiered pricing
- monthly invoice rollups
- invoice lifecycle foundation

### Payments
- Razorpay order creation
- HMAC-SHA256 webhook verification
- provider payment ID as idempotency boundary

### Observability
- OpenTelemetry traces
- Prometheus metrics
- structured logging
- request/correlation IDs

## Prerequisites

- Go 1.23+
- Docker
- Docker Compose
- protoc
- protoc-gen-go
- protoc-gen-go-grpc

Install protobuf plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Start infrastructure

```bash
cp .env.example .env
docker compose up -d
```

Couchbase is exposed at `http://localhost:8091`.

Create the `api_monetization` bucket, an `app` scope, and these collections:

```text
api_keys
wallets
ledger
idempotency
usage
invoices
payments
```

Then run `migrations/001_indexes.sql` in Couchbase Query Workbench.

## Generate protobuf code

```bash
make proto
```

## Build and test

```bash
go mod tidy
go build ./...
go test ./...
go test -race ./...
```

## Run services

Use four terminals:

```bash
make run-wallet
make run-metering
make run-billing
make run-gateway
```

Gateway: `http://localhost:8080`

Wallet gRPC: `localhost:9090`

Metering metrics: `http://localhost:9101/metrics`

Billing metrics: `http://localhost:9102/metrics`

Prometheus: `http://localhost:9099`

## Create an API key

```bash
curl -X POST http://localhost:8080/v1/keys \
  -H 'Content-Type: application/json' \
  -d '{"owner_id":"user_123","name":"dev","rate_limit":100}'
```

The plaintext key is returned only at creation time.

## Call the proxy

```bash
curl -X POST http://localhost:8080/v1/proxy/orders \
  -H 'X-API-Key: amp_<key>'
```

If `UPSTREAM_URL` is set, the gateway forwards to it. Otherwise it returns `202 Accepted` after the usage event is durably published.

## System-design claims

Use this wording in interviews:

> The system uses at-least-once event delivery and achieves exactly-once business effect through persistent idempotency and transactional state changes.

Do not claim that NATS itself gives application-level exactly-once processing.

## Production hardening checklist

- TLS/mTLS
- Kubernetes Secrets or external secret manager
- Redis HA/Sentinel/Cluster
- NATS accounts and permissions
- Couchbase HA/backups
- payment reconciliation job
- invoice finalization scheduler
- distributed tracing backend
- alerting
- chaos tests
- load tests
- API versioning
- authentication/authorization service
