package ratelimit

import (
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yesyoukenspace/go-ratelimit/internal/utils"
)

func TestAllow(t *testing.T) {
	rStr := utils.RandString(4)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	tests := []struct {
		name            string
		reqPerSec       float64
		burst           int
		runFor          time.Duration
		expectedAllowed int
		tolerance       float64
	}{
		{
			reqPerSec:       10,
			burst:           1,
			runFor:          3 * time.Second,
			expectedAllowed: 31,
			tolerance:       1,
		},
		{
			reqPerSec:       1,
			burst:           100,
			runFor:          3 * time.Second,
			expectedAllowed: 103,
			tolerance:       1,
		},
	}

	ratelimiters := []struct {
		name        string
		constructor func(float64, int) Ratelimiter
	}{
		{
			name: "SyncMap + Load > LoadOrStore",
			constructor: func(l float64, i int) Ratelimiter {
				return NewSyncMapLoadThenLoadOrStore(nil, l, i)
			},
		},
		{
			name: "SyncMap + Load > Store",
			constructor: func(l float64, i int) Ratelimiter {
				return NewSyncMapLoadThenStore(nil, l, i)
			},
		},
		{
			name: "SyncMap + LoadOrStore",
			constructor: func(l float64, i int) Ratelimiter {
				return NewSyncMapLoadOrStore(nil, l, i)
			},
		},
		{
			name: "Map + Mutex",
			constructor: func(l float64, i int) Ratelimiter {
				return NewMutex(nil, l, i)
			},
		},
		{
			name: "Map + RWMutex",
			constructor: func(l float64, i int) Ratelimiter {
				return NewRWMutex(nil, l, i)
			},
		},
		{
			name: "Redis",
			constructor: func(l float64, i int) Ratelimiter {
				return NewGoRedis(rdb, l, i)
			},
		},
	}
	for _, ratelimiter := range ratelimiters {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("ratelimiter=%s;rps=%2f;burst=%d", ratelimiter.name, tt.reqPerSec, tt.burst), func(t *testing.T) {
				allowed := 0
				denied := 0
				ticker := time.NewTicker(time.Millisecond)
				timer := time.NewTimer(tt.runFor)
				limiterGroup := ratelimiter.constructor(tt.reqPerSec, tt.burst)
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
					t.Errorf("expected >%d runs, got %d", int(tt.reqPerSec)*int(tt.runFor.Seconds()), allowed+denied)
				}
			})
		}
	}
}
