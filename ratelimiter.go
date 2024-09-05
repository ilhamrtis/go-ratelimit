package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

type Limiter interface {
	Allow() bool
}

type Ratelimiter interface {
	Allow(string) bool
}

type SyncMapLimiterWithLoadLoadStore struct {
	M sync.Map
	R rate.Limit
	B int
}

var _ Ratelimiter = &SyncMapLimiterWithLoadLoadStore{}

func (d *SyncMapLimiterWithLoadLoadStore) Allow(key string) bool {
	l, ok := d.M.Load(key)
	if !ok {
		l, _ = d.M.LoadOrStore(key, rate.NewLimiter(d.R, d.B))
	}
	return l.(Limiter).Allow()
}

type SyncMapLimiterWithLoadStore struct {
	M sync.Map
	R rate.Limit
	B int
}

func (d *SyncMapLimiterWithLoadStore) Allow(key string) bool {
	l, _ := d.M.LoadOrStore(key, rate.NewLimiter(d.R, d.B))
	return l.(Limiter).Allow()
}

var _ Ratelimiter = &SyncMapLimiterWithLoadStore{}

type DefaultLimiter = SyncMapLimiterWithLoadLoadStore

func NewDefaultLimiter(reqPerSec rate.Limit, burst int) *DefaultLimiter {
	return &DefaultLimiter{R: reqPerSec, B: burst}
}
