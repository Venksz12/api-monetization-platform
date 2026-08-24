# Razorpay Integration

Webhook processing must use the raw request body and the provider's signature.

Flow:

```text
Razorpay
   |
   v
Webhook endpoint
   |
   +-- verify HMAC signature
   |
   +-- extract payment/event ID
   |
   +-- check idempotency record
   |
   +-- update payment state
   |
   +-- credit wallet once
```

Never credit a wallet solely because an HTTP webhook request was received. Verify authenticity and make the credit idempotent.
