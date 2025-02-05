package ratelimit

import (
	"context"
	"fmt"
	"time"

	"testing"
)

func BenchmarkDistributed(b *testing.B) {
	rate := (100.0)
	burst := 100
	ratelimiters := []struct {
		name    string
		limiter Ratelimiter
	}{
		{name: "Redis", limiter: NewGoRedis(newRDB(2), rate, burst)},
		{name: "Redis with delay in sync", limiter: NewRedisWithDelay(context.Background(), RedisWithDelayOption{
			RedisClient:   newRDB(3),
			RequestPerSec: rate,
			Burst:         burst,
			SyncInterval:  time.Second / 2,
		})},
	}

	for _, concurrentUsers := range []int{2048, 16384} {
		for _, limiter := range ratelimiters {
			fmt.Println("Benchmarking", limiter.name, "with", concurrentUsers, "concurrent users")
		}
	}
}
