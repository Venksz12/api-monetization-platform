package gateway

import (
	"io"
	"net/http"

	"github.com/yourusername/api-monetization-platform/internal/payments"
)

func RazorpayWebhookHandler(secret string, onPayment func(http.ResponseWriter, *http.Request, payments.Webhook)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2<<20))
		if err != nil {
			http.Error(w, "invalid body", 400)
			return
		}
		if !payments.VerifyWebhook(body, r.Header.Get("X-Razorpay-Signature"), secret) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		event, err := payments.ParseWebhook(body)
		if err != nil {
			http.Error(w, "invalid webhook", 400)
			return
		}
		onPayment(w, r, event)
	})
}
