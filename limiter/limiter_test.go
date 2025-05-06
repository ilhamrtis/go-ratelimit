package limiter

import (
	"fmt"
	"testing"
	"time"

	"github.com/yesyoukenspace/go-ratelimit/internal/utils"
)

var limiters = []struct {
	name        string
	constructor func(float64, int) Limiter
}{
	{
		name: "golang.org/x/time/rate",
		constructor: func(limit float64, burst int) Limiter {
			return NewBuiltinLimiter(limit, burst)
		},
	},
	{
		name: "Bucket",
		constructor: func(limit float64, burst int) Limiter {
			return NewBucket(limit, burst)
		},
	},
	{
		name: "ResetbasedLimiter",
		constructor: func(limit float64, burst int) Limiter {
			return NewResetbasedLimiter(limit, burst)
		},
	},
}

func TestLimiterAllow(t *testing.T) {
	tests := []struct {
		name            string
		reqPerSec       float64
		burst           int
		runFor          time.Duration
		expectedAllowed int
		tolerance       float64
	}{
		{
			reqPerSec:       100,
			burst:           10,
			runFor:          3 * time.Second,
			expectedAllowed: 310,
			tolerance:       0.001,
		},
		{
			reqPerSec:       10,
			burst:           100,
			runFor:          3 * time.Second,
			expectedAllowed: 130,
			tolerance:       0.001,
		},
		{
			reqPerSec:       500,
			burst:           1000,
			runFor:          3 * time.Second,
			expectedAllowed: 2500,
			tolerance:       0.0001,
		},
	}

	for _, limiter := range limiters {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("limiter=%s;rps=%2f;burst=%d", limiter.name, tt.reqPerSec, tt.burst), func(t *testing.T) {
				t.Parallel()
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
				if !utils.IsCloseEnough(float64(tt.expectedAllowed), float64(allowed), tt.tolerance) {
					t.Errorf("expected %d, got %d", tt.expectedAllowed, allowed)
				}
				if allowed+denied < int(tt.reqPerSec)*int(tt.runFor.Seconds()) {
					t.Errorf("expected >%d runs, got %d", int(tt.reqPerSec)*int(tt.runFor.Seconds()), allowed+denied)
				}
				if denied < 1 {
					t.Errorf("expected >%d denials, got %d", 1, denied)
				}
			})
		}
	}
}
