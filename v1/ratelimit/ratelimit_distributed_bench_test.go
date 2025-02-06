package ratelimit

import (
	"time"

	"testing"
)

func BenchmarkDistributedRatelimiters(b *testing.B) {
	rate := (100.0)
	burst := 100
	ratelimiters := []struct {
		name    string
		limiter Ratelimiter
	}{
		{name: "Redis", limiter: NewGoRedis(newRDB(2), rate, burst)},
		{name: "Redis with delay in sync", limiter: NewRedisDelayedSync(RedisDelayedSyncOption{
			RedisClient:  newRDB(3),
			TokenPerSec:  rate,
			Burst:        burst,
			SyncInterval: time.Second / 2,
		})},
	}

	for _, concurrentUsers := range []int{2048, 16384} {
		for _, limiter := range ratelimiters {
			benchmarkRunner(b, limiter.name, concurrentUsers, limiter.limiter)
		}
	}
}
