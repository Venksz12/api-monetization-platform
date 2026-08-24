package payments

import (
	"encoding/json"
	"fmt"
)

type Webhook struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID       string `json:"id"`
				OrderID  string `json:"order_id"`
				Amount   int64  `json:"amount"`
				Currency string `json:"currency"`
				Status   string `json:"status"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

func ParseWebhook(body []byte) (Webhook, error) {
	var w Webhook
	if err := json.Unmarshal(body, &w); err != nil {
		return w, err
	}
	if w.Payload.Payment.Entity.ID == "" {
		return w, fmt.Errorf("payment id missing")
	}
	return w, nil
}
