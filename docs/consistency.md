# Consistency

## Wallet

A debit is a Couchbase multi-document transaction:

1. Read idempotency document.
2. Read wallet.
3. Validate currency and available balance.
4. Update wallet.
5. Insert immutable ledger entry.
6. Insert idempotency result.
7. Commit.

A retry with the same idempotency key returns the original ledger result.

## Event processing

Consumers ACK only after persistence succeeds. A crash before ACK causes redelivery. Persistent business idempotency prevents duplicate effects.

## Payments

Razorpay provider payment ID is the idempotency boundary. A webhook is accepted only after HMAC-SHA256 verification and duplicate provider IDs must not create another wallet credit.

## Monetary representation

Amounts are integer minor units. Never use floating-point values for money.
