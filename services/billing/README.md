# Billing

Consumes the same usage stream independently from metering.

This demonstrates fan-out through independent durable consumers. In production, billing should use an invoice-period aggregate, idempotency record, and wallet debit command.
