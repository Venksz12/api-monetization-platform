# API Monetization Platform

A distributed API monetization platform built with **Go**, using **NATS JetStream**, **Couchbase**, **Redis**, **gRPC**, **Razorpay**, **OpenTelemetry**, and **Prometheus**.

The platform authenticates API consumers, applies distributed rate limits, records API usage asynchronously, aggregates monthly usage, generates billing records, and provides a transactional wallet service over gRPC.

## Architecture

```text
                        ┌─────────────────┐
                        │      Client     │
                        └────────┬────────┘
                                 │
                              API Key
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │      API Gateway       │
                    │                        │
                    │ API Key Authentication │
                    │ Redis Rate Limiting    │
                    │ Request IDs            │
                    │ Correlation IDs        │
                    │ Reverse Proxy          │
                    └───────────┬────────────┘
                                │
                    usage.recorded event
                                │
                                ▼
                    ┌─────────────────────┐
                    │   NATS JetStream    │
                    └──────────┬──────────┘
                               │
                     ┌─────────┴─────────┐
                     │                   │
                     ▼                   ▼
              ┌──────────────┐    ┌──────────────┐
              │   Metering   │    │   Billing    │
              │   Consumer   │    │   Consumer   │
              └──────┬───────┘    └──────┬───────┘
                     │                   │
                     ▼                   ▼
                Couchbase           Couchbase
                  Usage              Invoices


                    ┌─────────────────────┐
                    │    Wallet gRPC      │
                    │       :9090         │
                    └──────────┬──────────┘
                               │
                               ▼
                         Couchbase
                       ┌──────┼───────┐
                       │      │       │
                    Wallet  Ledger  Idempotency

                    Razorpay Integration
                            │
                            ▼
                     Webhook Verification
```

## Features

### API Gateway

* API key authentication.
* API keys generated using cryptographically secure random bytes.
* SHA-256 API-key hashes stored in Couchbase.
* Redis-backed distributed token-bucket rate limiting.
* Configurable request rate and burst capacity.
* HTTP `429` responses when the rate limit is exceeded.
* `X-Request-ID` generation.
* `X-Correlation-ID` generation.
* Reverse proxy support through `UPSTREAM_URL`.
* Usage event publication to NATS JetStream.
* Prometheus metrics.
* OpenTelemetry HTTP instrumentation.

### Usage Metering

Usage events are published on:

```text
usage.recorded
```

The metering service uses a durable NATS JetStream pull consumer.

Processing flow:

```text
usage.recorded
      │
      ▼
Metering Consumer
      │
      ▼
Couchbase Usage Rollup
      │
      ▼
ACK
```

The consumer includes:

* Explicit acknowledgement after successful processing.
* Exponential redelivery delay.
* Maximum delivery attempts.
* Dead-letter publication after repeated failures.
* Monthly usage aggregation.

Usage documents use:

```text
usage::<owner_id>::<YYYY-MM>
```

Example:

```json
{
  "owner_id": "user_123",
  "period": "2026-08",
  "total_units": 1000,
  "total_requests": 25
}
```

### Billing

The billing worker consumes the same usage stream through a separate durable consumer.

The pricing engine supports tiered pricing.

Current reference pricing:

```text
0 – 100,000 requests
    0 minor currency units / request

100,000 – 1,000,000 requests
    1 minor currency unit / request

1,000,000+
    0 minor currency units / request
```

The current implementation calculates the monthly usage cost and persists an invoice document.

Invoice identifiers use:

```text
invoice::<owner_id>::<YYYY-MM>
```

Current invoice records are persisted with the initial:

```text
DRAFT
```

status.

### Wallet gRPC Service

The wallet service exposes:

```text
GetBalance
Credit
Debit
```

through:

```text
localhost:9090
```

Money is stored as integer minor units.

Example:

```text
100000 INR minor units = ₹1,000.00
```

Wallet mutations use Couchbase transactions involving:

```text
wallet
ledger entry
idempotency record
```

The ledger records immutable transaction entries.

Repeated wallet mutations using the same idempotency key return the existing transaction instead of applying the operation again.

The wallet protects against insufficient funds before applying a debit.

### Razorpay

The payment package currently provides:

* Razorpay order creation.
* Razorpay HTTP API client.
* HMAC-SHA256 webhook signature verification.
* Webhook payload validation.
* Payment ID extraction.

The gateway exposes a Razorpay webhook endpoint at:

```text
/v1/webhooks/razorpay
```

Verified webhook events are currently logged by the gateway.

### Observability

The services include:

* OpenTelemetry tracing initialization.
* Prometheus metrics.
* Structured logging with Zap.
* Request IDs.
* Correlation IDs.

Available metrics endpoints:

```text
Gateway
http://localhost:8080/metrics

Metering
http://localhost:9101/metrics

Billing
http://localhost:9102/metrics
```

## Technology Stack

| Component                 | Technology      |
| ------------------------- | --------------- |
| Language                  | Go              |
| External API              | HTTP            |
| Internal RPC              | gRPC / Protobuf |
| Event Streaming           | NATS JetStream  |
| Database                  | Couchbase       |
| Distributed Rate Limiting | Redis           |
| Payments                  | Razorpay        |
| Tracing                   | OpenTelemetry   |
| Metrics                   | Prometheus      |
| Logging                   | Zap             |
| Containers                | Docker          |
| Deployment                | Kubernetes      |

## Repository Structure

```text
api-monetization-platform/
│
├── cmd/
│   ├── gateway/
│   ├── wallet/
│   ├── metering/
│   └── billing/
│
├── internal/
│   ├── auth/
│   ├── config/
│   ├── events/
│   ├── ledger/
│   ├── messaging/
│   ├── observability/
│   ├── payments/
│   ├── pricing/
│   ├── ratelimit/
│   └── repository/
│
├── services/
│   ├── gateway/
│   ├── wallet/
│   ├── metering/
│   └── billing/
│
├── proto/
│   └── wallet.proto
│
├── migrations/
├── tests/
│   ├── concurrency/
│   ├── integration/
│   └── unit/
│
├── deployments/
│   ├── docker/
│   ├── kubernetes/
│   └── prometheus/
│
├── docs/
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Local Requirements

Install:

* Go 1.23+
* Docker Desktop
* Docker Compose
* Protocol Buffers compiler (`protoc`)
* `protoc-gen-go`
* `protoc-gen-go-grpc`
* `grpcurl` for convenient gRPC testing

Install protobuf generators:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Configuration

Create a local environment file:

```bash
cp .env.example .env
```

Important configuration:

```env
GATEWAY_HTTP_ADDR=:8080
WALLET_GRPC_ADDR=:9090

NATS_URL=nats://localhost:4222
NATS_STREAM=API_PLATFORM

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

COUCHBASE_CONNSTR=couchbase://localhost
COUCHBASE_USERNAME=Administrator
COUCHBASE_PASSWORD=password
COUCHBASE_BUCKET=api_monetization
COUCHBASE_SCOPE=app

RAZORPAY_KEY_ID=
RAZORPAY_KEY_SECRET=
RAZORPAY_WEBHOOK_SECRET=
```

Do not commit `.env` or payment credentials.

## Start Infrastructure

```bash
docker compose up -d
```

Local infrastructure:

```text
NATS        localhost:4222
Redis       localhost:6379
Couchbase   localhost:8091
Prometheus  localhost:9099
```

Create the Couchbase bucket:

```text
api_monetization
```

Create scope:

```text
app
```

Create the following collections:

```text
api_keys
wallets
ledger
idempotency
usage
invoices
payments
```

Then run:

```text
migrations/001_indexes.sql
```

from Couchbase Query Workbench.

## Generate Protobuf Code

Run:

```bash
make proto
```

This generates:

```text
proto/wallet.pb.go
proto/wallet_grpc.pb.go
```

## Build and Test

```bash
go mod tidy
go build ./...
go test ./...
go test -race ./...
```

The concurrency test covers:

* 100 concurrent wallet operations.
* Concurrent credits and debits.
* Idempotency replay.
* Final balance reconciliation.

## Run the Services

Start each service separately.

### Wallet

```bash
make run-wallet
```

### Metering

```bash
make run-metering
```

### Billing

```bash
make run-billing
```

### Gateway

```bash
make run-gateway
```

Endpoints:

```text
Gateway HTTP       :8080
Wallet gRPC        :9090
Metering metrics   :9101
Billing metrics    :9102
Prometheus         :9099
```

## API Usage

### Create an API Key

```bash
curl -X POST http://localhost:8080/v1/keys \
  -H "Content-Type: application/json" \
  -d '{
    "owner_id": "user_123",
    "name": "development",
    "rate_limit": 100
  }'
```

Example response:

```json
{
  "id": "key_...",
  "api_key": "amp_..."
}
```

The plaintext API key is returned when created. The stored representation is its SHA-256 hash.

### Proxy a Request

```bash
curl -X POST http://localhost:8080/v1/proxy/orders \
  -H "X-API-Key: amp_<YOUR_KEY>"
```

Optional usage units:

```bash
curl -X POST http://localhost:8080/v1/proxy/orders \
  -H "X-API-Key: amp_<YOUR_KEY>" \
  -H "X-API-Units: 10"
```

If `UPSTREAM_URL` is configured, the gateway forwards the request to that upstream service.

## Wallet Examples

### Get Balance

```bash
grpcurl -plaintext \
  -d '{"wallet_id":"wallet_123"}' \
  localhost:9090 \
  wallet.v1.WalletService/GetBalance
```

### Credit

```bash
grpcurl -plaintext \
  -d '{
    "wallet_id":"wallet_123",
    "amount_minor":100000,
    "currency":"INR",
    "idempotency_key":"credit-001",
    "reference":"initial-credit"
  }' \
  localhost:9090 \
  wallet.v1.WalletService/Credit
```

### Debit

```bash
grpcurl -plaintext \
  -d '{
    "wallet_id":"wallet_123",
    "amount_minor":25000,
    "currency":"INR",
    "idempotency_key":"debit-001",
    "reference":"api-usage"
  }' \
  localhost:9090 \
  wallet.v1.WalletService/Debit
```

Repeating the same mutation with the same idempotency key returns the previously created transaction instead of applying the mutation twice.

## Reliability Model

The event pipeline uses at-least-once delivery.

A consumer ACKs only after its handler succeeds:

```text
NATS
  │
  ▼
Consumer
  │
  ▼
Couchbase
  │
  ├── success ──► ACK
  │
  └── failure ──► NAK / retry
                       │
                       ▼
                      DLQ
```

The wallet mutation path uses persistent idempotency inside the transaction boundary so repeated commands do not create repeated financial effects.

## Deployment Files

Dockerfiles are available under:

```text
deployments/docker/
```

Kubernetes manifests are available under:

```text
deployments/kubernetes/
```

Prometheus configuration is available under:

```text
deployments/prometheus/prometheus.yml
```

## Current Implementation Scope

The repository currently implements the core API gateway, usage pipeline, billing persistence, transactional wallet foundation, Razorpay order/webhook verification, observability, tests, and deployment configuration described above.

Payment webhook verification is implemented, but the current webhook handler does not yet perform a persistent payment-state transition or wallet credit.

The current billing worker persists usage-derived invoices in `DRAFT` state; automated invoice finalization and payment reconciliation are not part of the current implementation.

## Engineering Principle

The project is designed around an explicit separation between:

```text
Synchronous API path
        │
        ▼
Authentication + Rate Limiting + Proxy
        │
        ▼
Asynchronous Usage Pipeline
        │
        ▼
Metering + Billing
        │
        ▼
Persistent Financial State
```

This keeps request handling independent from asynchronous billing work while providing durable storage and retry behavior for downstream processing.
