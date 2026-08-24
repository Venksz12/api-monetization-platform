package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	HTTPDuration      *prometheus.HistogramVec
	RateLimitRejects  *prometheus.CounterVec
	UsagePublished    prometheus.Counter
	EventsProcessed   *prometheus.CounterVec
	TransactionErrors prometheus.Counter
	ConsumerLag       prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "gateway_http_request_duration_seconds", Help: "Gateway HTTP request duration.",
		}, []string{"method", "route", "status"}),
		RateLimitRejects: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gateway_rate_limit_rejections_total", Help: "Rate-limit rejections.",
		}, []string{"key"}),
		UsagePublished: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gateway_usage_events_published_total", Help: "Usage events successfully published.",
		}),
		EventsProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "consumer_events_processed_total", Help: "Events processed by consumers.",
		}, []string{"consumer", "result"}),
		TransactionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "wallet_transaction_errors_total", Help: "Wallet transaction failures.",
		}),
		ConsumerLag: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "consumer_lag_messages", Help: "Approximate consumer lag.",
		}),
	}
	reg.MustRegister(m.HTTPDuration, m.RateLimitRejects, m.UsagePublished, m.EventsProcessed, m.TransactionErrors, m.ConsumerLag)
	return m
}

func Handler(reg prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

func ObserveHTTP(m *Metrics, start time.Time, method, route string, status int) {
	m.HTTPDuration.WithLabelValues(method, route, strconv.Itoa(status)).Observe(time.Since(start).Seconds())
}
