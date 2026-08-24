# Failure Handling

| Failure | Behavior |
|---|---|
| Redis unavailable | Gateway fails closed for protected routes |
| NATS unavailable | Gateway returns 503 rather than claiming usage was accepted |
| Consumer transient error | NAK with exponential delay |
| Consumer repeatedly fails | Publish to DLQ and terminate delivery |
| Consumer crashes after write before ACK | NATS redelivery + idempotent persistence |
| Wallet transaction conflict | Couchbase transaction retry/abort; caller can retry command |
| Insufficient funds | gRPC `FailedPrecondition` |
| Razorpay webhook duplicate | Provider payment ID idempotency |
| Upstream timeout | Gateway returns 502/timeout; usage event already records attempted request |
| Process shutdown | signal-aware context, NATS drain, gRPC graceful stop, HTTP shutdown |

Do not claim transport-level exactly-once semantics. Claim exactly-once business effect where persistent idempotency makes retries safe.
