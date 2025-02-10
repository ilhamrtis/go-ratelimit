package ratelimit

import (
	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type NewLimiterFn = func(rate float64, burst int) limiter.Limiter

type Ratelimiter interface {
	Allow(string) (bool, error)
	AllowN(string, int) (bool, error)
}

func NewDefaultLimiter(reqPerSec float64, burst int) limiter.Limiter {
	return limiter.NewResetbasedLimiter(reqPerSec, burst)
}
