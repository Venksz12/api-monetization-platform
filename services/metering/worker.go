package metering

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"

	"github.com/yourusername/api-monetization-platform/internal/events"
	"github.com/yourusername/api-monetization-platform/internal/messaging"
	"github.com/yourusername/api-monetization-platform/internal/observability"
	"github.com/yourusername/api-monetization-platform/internal/repository"
)

type Worker struct {
	Messaging *messaging.Client
	Usage     *repository.UsageRepository
	Metrics   *observability.Metrics
}

func (w *Worker) Run(ctx context.Context) error {
	c := &messaging.Consumer{
		JS: w.Messaging.JS, Stream: "API_PLATFORM",
		Subject: events.UsageRecordedSubject, Durable: "metering",
		MaxDeliver: 5, DLQSubject: events.UsageDLQSubject,
	}
	return c.Run(ctx, w.handle, func(ctx context.Context, subject string, data []byte) error {
		_, err := w.Messaging.JS.Publish(subject, data, nats.Context(ctx))
		return err
	})
}

func (w *Worker) handle(ctx context.Context, msg *nats.Msg) error {
	var env events.Envelope
	if err := json.Unmarshal(msg.Data, &env); err != nil {
		return err
	}
	var usage events.UsageRecorded
	if err := json.Unmarshal(env.Payload, &usage); err != nil {
		return err
	}
	period := usage.OccurredAt.UTC().Format("2006-01")
	if err := w.Usage.Add(ctx, usage.OwnerID, period, usage.Units); err != nil {
		w.Metrics.EventsProcessed.WithLabelValues("metering", "error").Inc()
		return err
	}
	w.Metrics.EventsProcessed.WithLabelValues("metering", "success").Inc()
	return nil
}
