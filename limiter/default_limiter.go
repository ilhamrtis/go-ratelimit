package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

type DefaultLimiter struct {
	*rate.Limiter
}

func NewDefaultLimiter(limit float64, burst int) *DefaultLimiter {
	return &DefaultLimiter{
		Limiter: rate.NewLimiter(rate.Limit(limit), burst),
	}
}

func (d *DefaultLimiter) AllowN(n int) (bool, error) {
	return d.Limiter.AllowN(time.Now(), n), nil
}

func (d *DefaultLimiter) Allow() (bool, error) {
	return d.Limiter.Allow(), nil
}

func (d *DefaultLimiter) ForceN(n int) (bool, error) {
	return d.Limiter.ReserveN(time.Now(), n).OK(), nil
}
