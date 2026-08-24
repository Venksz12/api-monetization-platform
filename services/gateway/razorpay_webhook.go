package main

// This file shows the handler boundary for Razorpay.
// Wire it into the gateway mux when the webhook secret is configured.
//
// Verify the raw request body with payments.VerifyWebhook before unmarshalling
// and before applying any wallet mutation.
//
// For production, persist the Razorpay event/payment ID as an idempotency key.
