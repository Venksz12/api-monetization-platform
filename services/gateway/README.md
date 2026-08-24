# Gateway

The gateway is intentionally small enough to understand but demonstrates the core flow:

1. Authenticate API key.
2. Apply per-key token bucket.
3. Create immutable usage event ID.
4. Publish to JetStream.
5. Return an accepted response.

For production, route the request to an upstream service after the usage event is accepted or use an outbox/event publication strategy when upstream execution and metering must share a stronger consistency boundary.
