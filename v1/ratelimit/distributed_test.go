package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yesyoukenspace/go-ratelimit/internal/utils"
)

func TestDistributedAllow(t *testing.T) {
	tests := []testDistributedRatelimiterConfig{
		{
			reqPerSec:       100,
			burst:           100,
			runFor:          20 * time.Second,
			expectedAllowed: 2100,
			tolerance:       0.02,
			instances:       5,
		},
		{
			reqPerSec:       100,
			burst:           100,
			runFor:          20 * time.Second,
			expectedAllowed: 2100,
			tolerance:       0.1,
			instances:       20,
		},
	}

	ratelimiters := []struct {
		name        string
		constructor func(float64, int) Ratelimiter
	}{
		{
			name: "Redis",
			constructor: func(l float64, i int) Ratelimiter {
				return NewGoRedis(newRDB(0), l, i)
			},
		},
		{
			name: "Redis with delay in sync",
			constructor: func(l float64, i int) Ratelimiter {
				return NewRedisDelayedSync(context.Background(), RedisDelayedSyncOption{
					redisClient:          newRDB(1),
					replenishedPerSecond: l,
					burst:                i,
					syncInterval:         time.Second / 10,
				})
			},
		},
	}
	for _, ratelimiter := range ratelimiters {
		for _, tt := range tests {
			testDistributedRatelimiter(t, testDistributedRatelimiterConfig{
				name:            ratelimiter.name,
				reqPerSec:       tt.reqPerSec,
				burst:           tt.burst,
				runFor:          tt.runFor,
				constructor:     ratelimiter.constructor,
				expectedAllowed: tt.expectedAllowed,
				tolerance:       tt.tolerance,
				instances:       tt.instances,
			})
		}
	}
}

type testDistributedRatelimiterConfig struct {
	name            string
	reqPerSec       float64
	burst           int
	runFor          time.Duration
	constructor     func(float64, int) Ratelimiter
	expectedAllowed int
	tolerance       float64
	instances       int
}

func testDistributedRatelimiter(t *testing.T, tt testDistributedRatelimiterConfig) {
	t.Run(fmt.Sprintf("ratelimiter=%s;rps=%2f;burst=%d;instances=%d", tt.name, tt.reqPerSec, tt.burst, tt.instances), func(t *testing.T) {
		t.Parallel()
		rStr := utils.RandString(4)
		totalAllowed := 0
		totalDenied := 0
		ratelimiters := make([]Ratelimiter, tt.instances)
		resultsChan := make(chan []int, tt.instances)
		for i := 0; i < tt.instances; i++ {
			ratelimiters[i] = tt.constructor(tt.reqPerSec, tt.burst)
		}
		for _, ratelimiter := range ratelimiters {
			ticker := time.NewTicker(time.Millisecond)
			timer := time.NewTimer(tt.runFor)
			go func(ratelimiter Ratelimiter) {
				denied := 0
				allowed := 0

				for {
					select {
					case <-ticker.C:
						if ok, _ := ratelimiter.Allow(rStr); ok {
							allowed++
						} else {
							denied++
						}
					case <-timer.C:
						ticker.Stop()
						resultsChan <- []int{allowed, denied}
						return
					}
				}
			}(ratelimiter)
		}

		for range ratelimiters {
			results := <-resultsChan
			totalAllowed += results[0]
			totalDenied += results[1]
		}

		if !utils.IsCloseEnough(float64(tt.expectedAllowed), float64(totalAllowed), tt.tolerance) {
			t.Errorf("unexpected allowed: expected %d, got %d", tt.expectedAllowed, totalAllowed)
		}
		expectedTotalRequests := int(tt.reqPerSec) * int(tt.runFor.Seconds()) * tt.instances
		if totalAllowed+totalDenied < expectedTotalRequests {
			t.Errorf("unexpected total requests: expected >%d total requests, got %d", expectedTotalRequests, totalAllowed+totalDenied)
		}
		if totalDenied < 1 {
			t.Errorf("expected at least 1 denied request")
		}
	})
}
