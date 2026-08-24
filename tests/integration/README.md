# Integration tests

Start NATS and Couchbase with Docker Compose, then run service integration tests.

Recommended cases:

- gateway -> NATS -> metering
- gateway -> NATS -> billing
- wallet credit -> debit
- duplicate debit idempotency
- insufficient balance
