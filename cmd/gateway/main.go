package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/yourusername/api-monetization-platform/internal/auth"
	"github.com/yourusername/api-monetization-platform/internal/config"
	"github.com/yourusername/api-monetization-platform/internal/messaging"
	"github.com/yourusername/api-monetization-platform/internal/observability"
	"github.com/yourusername/api-monetization-platform/internal/payments"
	"github.com/yourusername/api-monetization-platform/internal/ratelimit"
	"github.com/yourusername/api-monetization-platform/internal/repository"
	"github.com/yourusername/api-monetization-platform/services/gateway"
)

func main() {
	cfg := config.Load()
	log, _ := zap.NewProduction()
	defer log.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := observability.InitTracer(ctx, "api-monetization-gateway")
	if err != nil {
		log.Fatal("tracer", zap.Error(err))
	}
	defer shutdownTracer(context.Background())

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

	limiter := ratelimit.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RateLimit, cfg.RateBurst)
	reg := prometheus.NewRegistry()
	metrics := observability.NewMetrics(reg)

	h := &gateway.Handler{
		Keys: &repository.APIKeyRepository{CB: cb}, Limiter: limiter,
		Messaging: nc, Upstream: cfg.UpstreamURL, Metrics: metrics,
		Log: log, HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
	authMW := (&auth.APIKeyMiddleware{Repo: h.Keys}).Authenticate

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := cb.Ping(r.Context()); err != nil {
			http.Error(w, "not ready", 503)
			return
		}
		w.WriteHeader(200)
	})
	mux.HandleFunc("/metrics", observability.Handler(reg))
	mux.HandleFunc("/v1/keys", h.CreateKey)
	mux.Handle("/v1/proxy/", authMW(http.HandlerFunc(h.Proxy)))
	mux.Handle("/v1/webhooks/razorpay", gateway.RazorpayWebhookHandler(cfg.RazorpayWebhookSecret, func(w http.ResponseWriter, r *http.Request, event payments.Webhook) {
		h.Log.Info("razorpay webhook verified", zap.String("event", event.Event), zap.String("payment_id", event.Payload.Payment.Entity.ID))
		w.WriteHeader(http.StatusOK)
	}))

	srv := &http.Server{Addr: cfg.GatewayHTTPAddr, Handler: otelhttp.NewHandler(mux, "gateway-http"), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Info("gateway listening", zap.String("addr", cfg.GatewayHTTPAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http", zap.Error(err))
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
