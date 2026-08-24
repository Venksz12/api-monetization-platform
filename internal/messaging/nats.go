package messaging

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	NC *nats.Conn
	JS nats.JetStreamContext
}

func Connect(url string) (*Client, error) {
	nc, err := nats.Connect(url, nats.Timeout(5*time.Second), nats.MaxReconnects(-1))
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
		Subjects: []string{
			"usage.>",
			"billing.>",
			"payment.>",
			"deadletter.>",
		},
		Storage:           nats.FileStorage,
		Retention:         nats.LimitsPolicy,
		MaxAge:            7 * 24 * time.Hour,
		MaxMsgsPerSubject: -1,
	})
	return err
}

func (c *Client) Close() {
	_ = c.NC.Drain()
	c.NC.Close()
}

func ConsumerName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
