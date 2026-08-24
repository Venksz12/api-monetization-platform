package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/yourusername/api-monetization-platform/internal/auth"
	"github.com/yourusername/api-monetization-platform/internal/events"
	"github.com/yourusername/api-monetization-platform/internal/messaging"
	"github.com/yourusername/api-monetization-platform/internal/observability"
	"github.com/yourusername/api-monetization-platform/internal/ratelimit"
	"github.com/yourusername/api-monetization-platform/internal/repository"
)

type Handler struct {
	Keys       *repository.APIKeyRepository
	Limiter    *ratelimit.Limiter
	Messaging  *messaging.Client
	Upstream   string
	Metrics    *observability.Metrics
	Log        *zap.Logger
	HTTPClient *http.Client
}

func (h *Handler) CreateKey(w http.ResponseWriter, r *http.Request) {
	var in struct {
		OwnerID   string `json:"owner_id"`
		Name      string `json:"name"`
		RateLimit int    `json:"rate_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.OwnerID == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if in.RateLimit <= 0 {
		in.RateLimit = h.Limiter.Rate
	}

	plain, hash, err := auth.GenerateAPIKey()
	if err != nil {
		http.Error(w, "key generation failed", http.StatusInternalServerError)
		return
	}
	rec := repository.APIKey{
		ID: "key_" + uuid.NewString(), OwnerID: in.OwnerID, Name: in.Name,
		Hash: hash, Status: "ACTIVE", RateLimit: in.RateLimit, CreatedAt: time.Now().UTC(),
	}
	if err := h.Keys.Create(r.Context(), rec); err != nil {
		h.Log.Error("create api key", zap.Error(err))
		http.Error(w, "storage error", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": rec.ID, "api_key": plain})
}

func (h *Handler) Proxy(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	defer func() { observability.ObserveHTTP(h.Metrics, start, r.Method, "/v1/proxy", status) }()

	keyID, _ := r.Context().Value(auth.APIKeyContextKey).(string)
	ownerID, _ := r.Context().Value(auth.OwnerContextKey).(string)

	allowed, err := h.Limiter.Allow(r.Context(), keyID)
	if err != nil {
		status = http.StatusServiceUnavailable
		http.Error(w, "rate limiter unavailable", status)
		return
	}
	if !allowed {
		h.Metrics.RateLimitRejects.WithLabelValues(keyID).Inc()
		status = http.StatusTooManyRequests
		http.Error(w, "rate limit exceeded", status)
		return
	}

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.NewString()
	}
	correlationID := r.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.NewString()
	}

	units := int64(1)
	if v := r.Header.Get("X-API-Units"); v != "" {
		if _, err := fmt.Sscan(v, &units); err != nil || units <= 0 {
			units = 1
		}
	}

	event := events.UsageRecorded{
		EventID: uuid.NewString(), APIKeyID: keyID, OwnerID: ownerID,
		Path: r.URL.Path, Units: units, OccurredAt: time.Now().UTC(),
	}
	env, err := events.NewEnvelope(event.EventID, "usage.recorded", "gateway", requestID, correlationID, event)
	if err != nil {
		status = http.StatusInternalServerError
		http.Error(w, "event error", status)
		return
	}
	if _, err := h.Messaging.Publish(r.Context(), events.UsageRecordedSubject, env); err != nil {
		h.Log.Error("publish usage", zap.Error(err))
		status = http.StatusServiceUnavailable
		http.Error(w, "usage pipeline unavailable", status)
		return
	}
	h.Metrics.UsagePublished.Inc()

	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("X-Correlation-ID", correlationID)

	if h.Upstream == "" {
		writeJSON(w, http.StatusAccepted, map[string]any{"status": "accepted", "event_id": event.EventID})
		return
	}

	targetPath := strings.TrimPrefix(r.URL.Path, "/v1/proxy")
	target := strings.TrimRight(h.Upstream, "/") + "/" + strings.TrimLeft(targetPath, "/")
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		status = http.StatusInternalServerError
		http.Error(w, "proxy request error", status)
		return
	}
	req.Header = r.Header.Clone()
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("X-Correlation-ID", correlationID)
	req.Header.Del("X-API-Key")

	transport := h.HTTPClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	resp, err := otelhttp.NewTransport(transport).RoundTrip(req)
	if err != nil {
		status = http.StatusBadGateway
		http.Error(w, "upstream unavailable", status)
		return
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
