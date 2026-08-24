package billing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/yourusername/api-monetization-platform/internal/events"
	"github.com/yourusername/api-monetization-platform/internal/messaging"
	"github.com/yourusername/api-monetization-platform/internal/observability"
	"github.com/yourusername/api-monetization-platform/internal/pricing"
	"github.com/yourusername/api-monetization-platform/internal/repository"
)

type Worker struct {
	Messaging *messaging.Client
	Usage     *repository.UsageRepository
	Invoices  *repository.InvoiceRepository
	Metrics   *observability.Metrics
}

func (w *Worker) Run(ctx context.Context) error {
	c := &messaging.Consumer{
		JS: w.Messaging.JS, Stream: "API_PLATFORM",
		Subject: events.UsageRecordedSubject, Durable: "billing",
		MaxDeliver: 5, DLQSubject: "deadletter.billing",
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
	rollup, err := w.Usage.Get(ctx, usage.OwnerID, period)
	if err != nil {
		return err
	}
	total := pricing.Cost(rollup.TotalUnits, pricing.DefaultPlan)
	now := time.Now().UTC()
	inv := repository.Invoice{
		ID:      "invoice::" + usage.OwnerID + "::" + period,
		OwnerID: usage.OwnerID, Period: period,
		Subtotal: total, Tax: 0, Total: total,
		Currency: "INR", Status: "DRAFT", CreatedAt: now, UpdatedAt: now,
	}
	if err := w.Invoices.Upsert(ctx, inv); err != nil {
		w.Metrics.EventsProcessed.WithLabelValues("billing", "error").Inc()
		return err
	}
	w.Metrics.EventsProcessed.WithLabelValues("billing", "success").Inc()
	return nil
}
