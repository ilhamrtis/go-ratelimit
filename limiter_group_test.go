package ratelimit_test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yesyoukenspace/ratelimit"
)

func TestLimiterGroupAllow(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	token := make([]byte, 4)
	rand.Read(token)
	randString := base64.StdEncoding.EncodeToString(token)
	tests := []struct {
		name            string
		reqPerSec       ratelimit.ReqPerSec
		burst           int
		runFor          time.Duration
		expectedAllowed int
		tolerance       float64
	}{
		{
			reqPerSec:       100,
			burst:           100,
			runFor:          5 * time.Second,
			expectedAllowed: 600,
			tolerance:       1,
		},
		{
			reqPerSec:       1,
			burst:           100,
			runFor:          5 * time.Second,
			expectedAllowed: 105,
			tolerance:       1,
		},
	}

	limiterGroups := []struct {
		name        string
		constructor func(ratelimit.ReqPerSec, int) ratelimit.LimiterGroup
	}{
		{
			name: "SyncMap + Load > LoadOrStore",
			constructor: func(l ratelimit.ReqPerSec, i int) ratelimit.LimiterGroup {
				return &ratelimit.LiGrSyncMapLoadThenLoadOrStore{R: l, B: i}
			},
		},
		{
			name: "SyncMap + Load > Store",
			constructor: func(l ratelimit.ReqPerSec, i int) ratelimit.LimiterGroup {
				return &ratelimit.LiGrSyncMapLoadThenStore{R: l, B: i}
			},
		},
		{
			name: "SyncMap + LoadOrStore",
			constructor: func(l ratelimit.ReqPerSec, i int) ratelimit.LimiterGroup {
				return &ratelimit.LiGrSyncMapLoadOrStore{R: l, B: i}
			},
		},
		{
			name: "Map + Mutex",
			constructor: func(l ratelimit.ReqPerSec, i int) ratelimit.LimiterGroup {
				return ratelimit.NewLiGrMutex(l, i)
			},
		},
		{
			name: "Map + RWMutex",
			constructor: func(l ratelimit.ReqPerSec, i int) ratelimit.LimiterGroup {
				return ratelimit.NewLiGrRWMutex(l, i)
			},
		},
		{
			name: "Redis",
			constructor: func(l ratelimit.ReqPerSec, i int) ratelimit.LimiterGroup {
				return ratelimit.NewLiGrRedis(rdb, l, i)
			},
		},
	}
	for _, limiterGroup := range limiterGroups[5:] {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("liGr=%s;rps=%2f;burst=%d", limiterGroup.name, tt.reqPerSec, tt.burst), func(t *testing.T) {
				allowed := 0
				denied := 0
				ticker := time.NewTicker(time.Millisecond)
				timer := time.NewTimer(tt.runFor)
				limiterGroup := limiterGroup.constructor(tt.reqPerSec, tt.burst)
			L:
				for {
					select {
					case <-ticker.C:
						if limiterGroup.Allow(randString) {
							allowed++
						} else {
							denied++
						}
					case <-timer.C:
						ticker.Stop()
						break L
					}
				}
				if !isCloseEnough(float64(tt.expectedAllowed), float64(allowed), tt.tolerance) {
					t.Errorf("expected %d, got %d", tt.expectedAllowed, allowed)
				}
				if allowed+denied < int(tt.reqPerSec)*int(tt.runFor.Seconds()) {
					t.Errorf("expected >%d runs, got %d", int(tt.reqPerSec)*int(tt.runFor.Seconds()), allowed+denied)
				}
			})
		}
	}
}
