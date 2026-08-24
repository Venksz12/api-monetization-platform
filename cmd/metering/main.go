package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/yourusername/api-monetization-platform/internal/config"
	"github.com/yourusername/api-monetization-platform/internal/messaging"
	"github.com/yourusername/api-monetization-platform/internal/observability"
	"github.com/yourusername/api-monetization-platform/internal/repository"
	"github.com/yourusername/api-monetization-platform/services/metering"
)

func main() {
	cfg := config.Load()
	log, _ := zap.NewProduction()
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cb, err := repository.Connect(ctx, cfg.CBConnStr, cfg.CBUsername, cfg.CBPassword, cfg.CBBucket, cfg.CBScope)
	if err != nil {
		log.Fatal("couchbase", zap.Error(err))
	}
	defer cb.Close()

	nc, err := messaging.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatal("nats", zap.Error(err))
	}
	defer nc.Close()
	if err := nc.EnsureStream(cfg.NATSStream); err != nil {
		log.Fatal("stream", zap.Error(err))
	}

	reg := prometheus.NewRegistry()
	metrics := observability.NewMetrics(reg)
	worker := &metering.Worker{
		Messaging: nc, Usage: &repository.UsageRepository{CB: cb}, Metrics: metrics,
	}
	go func() {
		if err := worker.Run(ctx); err != nil && ctx.Err() == nil {
			log.Fatal("worker", zap.Error(err))
		}
	}()

	http.Handle("/metrics", observability.Handler(reg))
	srv := &http.Server{Addr: cfg.MeteringMetricsAddr, Handler: http.DefaultServeMux}
	go srv.ListenAndServe()
	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
}
