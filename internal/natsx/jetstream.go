package natsx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	NC *nats.Conn
	JS nats.JetStreamContext
}

func Connect(url string) (*Client, error) {
	nc, err := nats.Connect(url, nats.Timeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &Client{NC: nc, JS: js}, nil
}

func (c *Client) EnsureStream(name string) error {
	_, err := c.JS.StreamInfo(name)
	if err == nil {
		return nil
	}
	_, err = c.JS.AddStream(&nats.StreamConfig{
		Name: name,
		Subjects: []string{"usage.>"},
		Storage: nats.FileStorage,
		MaxAge: 7 * 24 * time.Hour,
	})
	return err
}

func (c *Client) Publish(ctx context.Context, subject string, v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	msg, err := c.JS.Publish(subject, b, nats.Context(ctx))
	if err != nil {
		return "", err
	}
	return msg.Stream + ":" + msg.Sequence.String(), nil
}
