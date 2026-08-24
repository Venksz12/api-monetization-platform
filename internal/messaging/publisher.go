package messaging

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

func (c *Client) Publish(ctx context.Context, subject string, value any) (*nats.PubAck, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return c.JS.Publish(subject, b, nats.Context(ctx))
}
