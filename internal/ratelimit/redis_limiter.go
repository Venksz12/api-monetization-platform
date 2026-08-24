package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	Client *redis.Client
	Rate   int
	Burst  int
}

const script = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local burst = tonumber(ARGV[3])

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if tokens == nil then
  tokens = burst
  ts = now
end

local elapsed = math.max(0, now - ts)
tokens = math.min(burst, tokens + elapsed * rate)

local allowed = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "ts", now)
redis.call("EXPIRE", key, math.ceil((burst / rate) * 2) + 10)
return allowed
`

func New(addr, password string, db, rate, burst int) *Limiter {
	return &Limiter{
		Client: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db}),
		Rate:   rate, Burst: burst,
	}
}

func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	if l.Rate <= 0 {
		return true, nil
	}
	v, err := l.Client.Eval(ctx, script, []string{fmt.Sprintf("rl:%s", key)},
		time.Now().UnixNano(), float64(l.Rate)/1e9, l.Burst).Int()
	if err != nil {
		return false, err
	}
	return v == 1, nil
}
