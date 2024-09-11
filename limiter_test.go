package ratelimit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/yesyoukenspace/ratelimit"
	"golang.org/x/time/rate"
)

func TestLimiterAllow(t *testing.T) {
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

	limiters := []struct {
		name        string
		constructor func(ratelimit.ReqPerSec, int) ratelimit.Limiter
	}{
		{
			name: "golang.org/x/time/rate",
			constructor: func(limit ratelimit.ReqPerSec, burst int) ratelimit.Limiter {
				return rate.NewLimiter(rate.Limit(limit), burst)
			},
		},
		{
			name: "Bucket",
			constructor: func(limit ratelimit.ReqPerSec, burst int) ratelimit.Limiter {
				return ratelimit.NewBucket(limit, burst)
			},
		},
	}
	for _, limiter := range limiters {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("liGr=%s;rps=%2f;burst=%d", limiter.name, tt.reqPerSec, tt.burst), func(t *testing.T) {
				allowed := 0
				denied := 0
				ticker := time.NewTicker(time.Millisecond / 2)
				timer := time.NewTimer(tt.runFor)
				limiter := limiter.constructor(tt.reqPerSec, tt.burst)
			L:
				for {
					select {
					case <-ticker.C:
						if limiter.Allow() {
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
