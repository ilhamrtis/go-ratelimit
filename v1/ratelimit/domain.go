package ratelimit

import (
	"github.com/yesyoukenspace/go-ratelimit/limiter"
	"golang.org/x/time/rate"
)

type NewLimiterFn = func(rate float64, burst int) limiter.Limiter
type Ratelimiter interface {
	Allow(string) (bool, error)
}

func NewDefaultRatelimiter(reqPerSec float64, burst int) Ratelimiter {
	return &SyncMapLoadThenLoadOrStore{R: reqPerSec, B: burst}
}

func NewDefaultLimiter(reqPerSec float64, burst int) limiter.Limiter {
	return limiter.NewSimpleLimiterAdapter(rate.NewLimiter(rate.Limit(reqPerSec), burst))
}
