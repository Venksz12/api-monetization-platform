package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/yourusername/api-monetization-platform/internal/auth"
	"github.com/yourusername/api-monetization-platform/internal/config"
	"github.com/yourusername/api-monetization-platform/internal/events"
	"github.com/yourusername/api-monetization-platform/internal/natsx"
	"github.com/yourusername/api-monetization-platform/internal/observability"
)

type keyRecord struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	RateLimit int   `json:"rate_limit"`
}

type gateway struct {
	mu       sync.RWMutex
	keys     map[string]keyRecord
	limiters map[string]*rate.Limiter
	nats     *natsx.Client
	log      *zap.Logger
}

func main() {
	cfg := config.Load()
	log := observability.NewLogger()
	defer log.Sync()

	nc, err := natsx.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatal("nats", zap.Error(err))
	}
	defer nc.NC.Close()
	if err := nc.EnsureStream(cfg.NATSStream); err != nil {
		log.Fatal("stream", zap.Error(err))
	}

	g := &gateway{
		keys: map[string]keyRecord{
			"demo": {ID: "key_demo", OwnerID: "demo_user", Name: "demo", Hash: auth.Hash("demo"), RateLimit: 100},
		},
		limiters: make(map[string]*rate.Limiter),
		nats: nc,
		log: log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", g.health)
	mux.HandleFunc("/readyz", g.health)
	mux.HandleFunc("/v1/keys", g.createKey)
	mux.HandleFunc("/v1/proxy", g.proxy)

	srv := &http.Server{
		Addr: cfg.HTTPAddr, Handler: logging(mux, log),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Info("gateway started", zap.String("addr", cfg.HTTPAddr))
	log.Fatal(srv.ListenAndServe())
}

func (g *gateway) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (g *gateway) createKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var in struct {
		OwnerID string `json:"owner_id"`
		Name string `json:"name"`
		RateLimit int `json:"rate_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.OwnerID == "" {
		http.Error(w, "invalid body", 400)
		return
	}
	if in.RateLimit <= 0 {
		in.RateLimit = 100
	}
	plain, hash, err := auth.Generate()
	if err != nil {
		http.Error(w, "generation failed", 500)
		return
	}
	rec := keyRecord{ID: "key_" + uuid.NewString(), OwnerID: in.OwnerID, Name: in.Name, Hash: hash, RateLimit: in.RateLimit}
	g.mu.Lock()
	g.keys[hash] = rec
	g.mu.Unlock()
	writeJSON(w, 201, map[string]any{"id": rec.ID, "api_key": plain})
}

func (g *gateway) proxy(w http.ResponseWriter, r *http.Request) {
	key := auth.Normalize(r.Header.Get("X-API-Key"))
	if key == "" {
		http.Error(w, "missing api key", 401)
		return
	}
	hash := auth.Hash(key)
	g.mu.RLock()
	rec, ok := g.keys[hash]
	lim := g.limiters[hash]
	g.mu.RUnlock()
	if !ok {
		http.Error(w, "invalid api key", 401)
		return
	}
	if lim == nil {
		g.mu.Lock()
		lim = g.limiters[hash]
		if lim == nil {
			lim = rate.NewLimiter(rate.Limit(rec.RateLimit), rec.RateLimit)
			g.limiters[hash] = lim
		}
		g.mu.Unlock()
	}
	if !lim.Allow() {
		http.Error(w, "rate limit exceeded", 429)
		return
	}

	units := int64(1)
	if v := r.Header.Get("X-API-Units"); v != "" {
		if _, err := fmt.Sscan(v, &units); err != nil || units <= 0 {
			units = 1
		}
	}
	event := events.UsageRecorded{
		EventID: uuid.NewString(), APIKeyID: rec.ID, OwnerID: rec.OwnerID,
		Path: r.Header.Get("X-API-Path"), Units: units, OccurredAt: time.Now().UTC(),
	}
	if event.Path == "" {
		event.Path = "/unknown"
	}
	if _, err := g.nats.Publish(context.Background(), events.UsageSubject, event); err != nil {
		g.log.Error("publish usage", zap.Error(err))
		http.Error(w, "usage pipeline unavailable", 503)
		return
	}
	writeJSON(w, 200, map[string]any{"status": "accepted", "event_id": event.EventID})
}

func logging(next http.Handler, log *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("http request", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.Duration("duration", time.Since(start)))
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
