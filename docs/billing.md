# Billing Model

Billing is based on usage units.

The default example plan is:

- first 100,000 requests: free
- next 900,000 requests: 1 minor currency unit/request
- future tiers can be added

A real implementation should calculate invoice-period deltas rather than charging each request individually. That reduces ledger volume and makes reconciliation easier.
