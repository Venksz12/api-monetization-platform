package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/nats-io/nats.go"
)

type Handler func(context.Context, *nats.Msg) error

type Consumer struct {
	JS         nats.JetStreamContext
	Stream     string
	Subject    string
	Durable    string
	MaxDeliver int
	DLQSubject string
}

func (c *Consumer) Run(ctx context.Context, handler Handler, publishDLQ func(context.Context, string, []byte) error) error {
	sub, err := c.JS.PullSubscribe(c.Subject, c.Durable, nats.BindStream(c.Stream))
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msgs, err := sub.Fetch(10, nats.MaxWait(1*time.Second))
		if err != nil && !errors.Is(err, nats.ErrTimeout) {
			return err
		}

		for _, msg := range msgs {
			mctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := handler(mctx, msg)
			cancel()
			if err == nil {
				_ = msg.Ack()
				continue
			}

			md, mdErr := msg.Metadata()
			if mdErr != nil {
				_ = msg.Nak()
				continue
			}

			if c.MaxDeliver > 0 && int(md.NumDelivered) >= c.MaxDeliver {
				if publishDLQ != nil {
					_ = publishDLQ(ctx, c.DLQSubject, append([]byte(nil), msg.Data...))
				}
				_ = msg.Term()
				continue
			}

			// Exponential redelivery delay. JetStream's NakWithDelay gives a
			// deterministic backoff while retaining at-least-once delivery.
			attempt := int(md.NumDelivered)
			delay := time.Duration(math.Min(float64(30*time.Second), float64(time.Second)*math.Pow(2, float64(attempt-1))))
			_ = msg.NakWithDelay(delay)
		}
	}
}

func Decode[T any](msg *nats.Msg) (T, error) {
	var out T
	err := json.Unmarshal(msg.Data, &out)
	return out, err
}
