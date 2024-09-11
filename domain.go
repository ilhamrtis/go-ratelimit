package ratelimit

import (
	"golang.org/x/time/rate"
)

type ReqPerSec = float64

type Limiter interface {
	Allow() bool
}

type LimiterGroup interface {
	Allow(string) bool
}

func NewDefaultLiGr(reqPerSec ReqPerSec, burst int) LimiterGroup {
	return &LiGrSyncMapLoadThenLoadOrStore{R: reqPerSec, B: burst}
}

func NewDefaultLimiter(reqPerSec ReqPerSec, burst int) Limiter {
	return rate.NewLimiter(rate.Limit(reqPerSec), burst)
}
