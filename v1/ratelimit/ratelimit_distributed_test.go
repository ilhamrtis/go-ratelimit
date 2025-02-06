package ratelimit

import (
	"fmt"
	"testing"
	"time"

	"github.com/yesyoukenspace/go-ratelimit/internal/utils"
)

func TestDistributedAllow(t *testing.T) {
	tests := []testDistributedRatelimiterConfig{
		{
			reqPerSec:       100,
			burst:           200,
			runFor:          20 * time.Second,
			expectedAllowed: 2200,
			tolerance:       0.05,
			instances:       5,
		},
		{
			reqPerSec:       100,
			burst:           200,
			runFor:          20 * time.Second,
			expectedAllowed: 2200,
			tolerance:       0.01,
			instances:       3,
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
				return NewRedisDelayedSync(RedisDelayedSyncOption{
					RedisClient:  newRDB(1),
					TokenPerSec:  l,
					Burst:        i,
					SyncInterval: time.Second / 2,
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
	t.Run(fmt.Sprintf("ratelimiter=%s;rps=%2f;burst=%d", tt.name, tt.reqPerSec, tt.burst), func(t *testing.T) {
		t.Parallel()
		rStr := utils.RandString(4)
		totalAllowed := 0
		totalDenied := 0
		lgs := make([]Ratelimiter, tt.instances)
		resultsChan := make(chan []int, tt.instances)
		for i := 0; i < tt.instances; i++ {
			lgs[i] = tt.constructor(tt.reqPerSec, tt.burst)
		}
		for _, lg := range lgs {
			ticker := time.NewTicker(time.Millisecond)
			timer := time.NewTimer(tt.runFor)
			go func(lg Ratelimiter) {
				denied := 0
				allowed := 0

				for {
					select {
					case <-ticker.C:
						if ok, _ := lg.Allow(rStr); ok {
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
			}(lg)
		}

		for _, _ = range lgs {
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
	})
}
