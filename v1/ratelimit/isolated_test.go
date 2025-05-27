package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yesyoukenspace/go-ratelimit/internal/utils"
)

func TestIsolatedAllow(t *testing.T) {
	tests := []testRatelimiterConfig{
		{
			reqPerSec:       100,
			burst:           10,
			runFor:          3 * time.Second,
			expectedAllowed: 310,
			tolerance:       0.01,
		},
		{
			reqPerSec:       10,
			burst:           100,
			runFor:          3 * time.Second,
			expectedAllowed: 130,
			tolerance:       0.01,
		},
	}

	ratelimiters := []struct {
		name        string
		constructor func(float64, int) Ratelimiter
	}{
		{
			name: "SyncMap + Load > LoadOrStore",
			constructor: func(l float64, i int) Ratelimiter {
				return NewSyncMapLoadThenLoadOrStore(NewDefaultLimiter, l, i)
			},
		},
		{
			name: "SyncMap + Load > Store",
			constructor: func(l float64, i int) Ratelimiter {
				return NewSyncMapLoadThenStore(NewDefaultLimiter, l, i)
			},
		},
		{
			name: "SyncMap + LoadOrStore",
			constructor: func(l float64, i int) Ratelimiter {
				return NewSyncMapLoadOrStore(NewDefaultLimiter, l, i)
			},
		},
		{
			name: "Map + Mutex",
			constructor: func(l float64, i int) Ratelimiter {
				return NewMutex(NewDefaultLimiter, l, i)
			},
		},
		{
			name: "Map + RWMutex",
			constructor: func(l float64, i int) Ratelimiter {
				return NewRWMutex(NewDefaultLimiter, l, i)
			},
		},
		{
			name: "Redis",
			constructor: func(l float64, i int) Ratelimiter {
				return NewGoRedis(newRDB(), l, i)
			},
		},
		{
			name: "Redis with delay in sync",
			constructor: func(l float64, i int) Ratelimiter {
				return NewRedisDelayedSync(context.Background(), RedisDelayedSyncOption{
					redisClient:          newRDB(),
					replenishedPerSecond: l,
					burst:                i,
					syncInterval:         time.Second / 2,
				})
			},
		},
	}
	for _, ratelimiter := range ratelimiters {
		for _, tt := range tests {
			testRatelimiter(t, testRatelimiterConfig{
				name:            ratelimiter.name,
				reqPerSec:       tt.reqPerSec,
				burst:           tt.burst,
				runFor:          tt.runFor,
				constructor:     ratelimiter.constructor,
				expectedAllowed: tt.expectedAllowed,
				tolerance:       tt.tolerance,
			})
		}
	}
}

type testRatelimiterConfig struct {
	name            string
	reqPerSec       float64
	burst           int
	runFor          time.Duration
	constructor     func(float64, int) Ratelimiter
	expectedAllowed int
	tolerance       float64
}

func testRatelimiter(t *testing.T, tt testRatelimiterConfig) {
	t.Run(fmt.Sprintf("ratelimiter=%s;rps=%2f;burst=%d", tt.name, tt.reqPerSec, tt.burst), func(t *testing.T) {
		rStr := utils.RandString(4)
		t.Parallel()
		allowed := 0
		denied := 0
		ticker := time.NewTicker(time.Millisecond)
		timer := time.NewTimer(tt.runFor)
		limiterGroup := tt.constructor(tt.reqPerSec, tt.burst)
	L:
		for {
			select {
			case <-ticker.C:
				if ok, _ := limiterGroup.Allow(rStr); ok {
					allowed++
				} else {
					denied++
				}
			case <-timer.C:
				ticker.Stop()
				break L
			}
		}
		if !utils.IsCloseEnough(float64(tt.expectedAllowed), float64(allowed), tt.tolerance) {
			t.Errorf("expected %d, got %d", tt.expectedAllowed, allowed)
		}
		if allowed+denied < int(tt.reqPerSec)*int(tt.runFor.Seconds()) {
			t.Errorf("expected >%d total requests, got %d", int(tt.reqPerSec)*int(tt.runFor.Seconds()), allowed+denied)
		}
	})
}
