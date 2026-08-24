# Architecture

```text
Client
  |
  v
Gateway
  |-- API key auth
  |-- Redis distributed token bucket
  |-- request/correlation IDs
  |-- reverse proxy
  |
  +--> NATS JetStream: usage.recorded
                 |
                 +--> Metering durable consumer --> Couchbase usage rollup
                 |
                 +--> Billing durable consumer --> invoice state

Wallet gRPC --> Couchbase transactions --> wallet + immutable ledger + idempotency
Razorpay --> webhook verification --> payment idempotency --> wallet credit
```

## Reliability boundary

NATS provides durable at-least-once delivery. The application provides exactly-once business effect by using persistent idempotency records and immutable ledger transactions.

## Horizontal scaling

Gateway replicas share rate-limit state in Redis. Metering and billing replicas share durable NATS consumer state and compete for messages within the same durable consumer.

## Data ownership

- API keys: `api_keys`
- Wallets: `wallets`
- Ledger: `ledger`
- Idempotency: `idempotency`
- Usage: `usage`
- Invoices: `invoices`
- Payments: `payments`
