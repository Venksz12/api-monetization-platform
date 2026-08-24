# Wallet

The wallet service exposes gRPC methods:

- GetBalance
- Credit
- Debit

The in-memory ledger is deliberately dependency-light for the first milestone. The production milestone should replace the in-memory implementation with Couchbase transactional writes while preserving the same domain interface.
