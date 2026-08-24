package config

import "os"

type Config struct {
	GatewayHTTPAddr     string
	WalletGRPCAddr      string
	MeteringMetricsAddr string
	BillingMetricsAddr  string
	UpstreamURL         string

	NATSURL        string
	NATSStream     string
	NATSMaxDeliver int

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RateLimit     int
	RateBurst     int

	CBConnStr  string
	CBUsername string
	CBPassword string
	CBBucket   string
	CBScope    string

	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string
}

func Load() Config {
	return Config{
		GatewayHTTPAddr:       env("GATEWAY_HTTP_ADDR", ":8080"),
		WalletGRPCAddr:        env("WALLET_GRPC_ADDR", ":9090"),
		MeteringMetricsAddr:   env("METERING_METRICS_ADDR", ":9101"),
		BillingMetricsAddr:    env("BILLING_METRICS_ADDR", ":9102"),
		UpstreamURL:           os.Getenv("UPSTREAM_URL"),
		NATSURL:               env("NATS_URL", "nats://localhost:4222"),
		NATSStream:            env("NATS_STREAM", "API_PLATFORM"),
		NATSMaxDeliver:        envInt("NATS_MAX_DELIVER", 5),
		RedisAddr:             env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         os.Getenv("REDIS_PASSWORD"),
		RedisDB:               envInt("REDIS_DB", 0),
		RateLimit:             envInt("RATE_LIMIT_PER_SECOND", 100),
		RateBurst:             envInt("RATE_LIMIT_BURST", 100),
		CBConnStr:             env("COUCHBASE_CONNSTR", "couchbase://localhost"),
		CBUsername:            env("COUCHBASE_USERNAME", "Administrator"),
		CBPassword:            env("COUCHBASE_PASSWORD", "password"),
		CBBucket:              env("COUCHBASE_BUCKET", "api_monetization"),
		CBScope:               env("COUCHBASE_SCOPE", "app"),
		RazorpayKeyID:         os.Getenv("RAZORPAY_KEY_ID"),
		RazorpayKeySecret:     os.Getenv("RAZORPAY_KEY_SECRET"),
		RazorpayWebhookSecret: os.Getenv("RAZORPAY_WEBHOOK_SECRET"),
	}
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func envInt(k string, fallback int) int {
	v := os.Getenv(k)
	var n int
	if v == "" {
		return fallback
	}
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
