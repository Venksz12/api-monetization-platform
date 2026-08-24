package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	KeyID   string
	Secret  string
	BaseURL string
	HTTP    *http.Client
}

type Order struct {
	ID       string `json:"id"`
	Entity   string `json:"entity"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

func NewClient(keyID, secret string) *Client {
	return &Client{KeyID: keyID, Secret: secret, BaseURL: "https://api.razorpay.com/v1", HTTP: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) CreateOrder(ctx context.Context, amount int64, currency, receipt string) (Order, error) {
	var out Order
	body, err := json.Marshal(map[string]any{"amount": amount, "currency": currency, "receipt": receipt, "payment_capture": 1})
	if err != nil {
		return out, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/orders", bytes.NewReader(body))
	if err != nil {
		return out, err
	}
	req.SetBasicAuth(c.KeyID, c.Secret)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return out, fmt.Errorf("razorpay order status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

func VerifyWebhook(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
