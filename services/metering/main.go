package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/yourusername/api-monetization-platform/internal/config"
	"github.com/yourusername/api-monetization-platform/internal/events"
	"github.com/yourusername/api-monetization-platform/internal/natsx"
	"github.com/yourusername/api-monetization-platform/internal/observability"
)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger()
	defer logger.Sync()

	nc, err := natsx.Connect(cfg.NATSURL)
	if err != nil { logger.Fatal("nats", zap.Error(err)) }
	defer nc.NC.Close()
	if err := nc.EnsureStream(cfg.NATSStream); err != nil { logger.Fatal("stream", zap.Error(err)) }

	sub, err := nc.JS.PullSubscribe(events.UsageSubject, "metering-workers", nats.BindStream(cfg.NATSStream))
	if err != nil { logger.Fatal("subscribe", zap.Error(err)) }

	logger.Info("metering consumer started")
	for {
		msgs, err := sub.Fetch(20, nats.MaxWait(2*time.Second))
		if err != nil && err != nats.ErrTimeout {
			logger.Error("fetch", zap.Error(err))
			continue
		}
		for _, msg := range msgs {
			var event events.UsageRecorded
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				_ = msg.Term()
				continue
			}
			// Replace this log with durable usage aggregation in Couchbase.
			logger.Info("usage metered",
				zap.String("event_id", event.EventID),
				zap.String("owner_id", event.OwnerID),
				zap.Int64("units", event.Units))
			if err := msg.Ack(); err != nil {
				log.Printf("ack: %v", err)
			}
		}
	}
}

