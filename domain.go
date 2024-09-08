package ratelimit

import (
	"golang.org/x/time/rate"
)

type Limit = rate.Limit

type Limiter interface {
	Allow() bool
}

type LimiterGroup interface {
	Allow(string) bool
}

type DefaultLimiterGroup = LiGrSyncMapLoadThenLoadOrStore

func NewDefaultLimiterGroup(reqPerSec Limit, burst int) *DefaultLimiterGroup {
	return &DefaultLimiterGroup{R: reqPerSec, B: burst}
}

type DefaultLimiter = rate.Limiter

func NewDefaultLimiter(reqPerSec Limit, burst int) *DefaultLimiter {
	return rate.NewLimiter(reqPerSec, burst)
}
