package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

type BuiltinLimiter struct {
	*rate.Limiter
}

func NewBuiltinLimiter(limit float64, burst int) *BuiltinLimiter {
	return &BuiltinLimiter{
		Limiter: rate.NewLimiter(rate.Limit(limit), burst),
	}
}

func (d *BuiltinLimiter) AllowN(n int) bool {
	return d.Limiter.AllowN(time.Now(), n)
}

func (d *BuiltinLimiter) Allow() bool {
	return d.Limiter.Allow()
}

func (d *BuiltinLimiter) ForceN(n int) bool {
	return d.Limiter.ReserveN(time.Now(), n).OK()
}
