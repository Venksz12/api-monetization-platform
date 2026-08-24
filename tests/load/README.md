# Load tests

Use k6, vegeta, or another load generator.

Suggested acceptance test:

- 1,000 concurrent API clients
- 10 minutes
- p95 gateway latency < 100 ms excluding downstream business work
- zero accepted requests lost from JetStream
- no duplicate business charges
- bounded consumer lag
