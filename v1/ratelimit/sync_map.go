package ratelimit

import (
	"sync"

	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type SyncMapLoadThenLoadOrStore[Limiter limiter.Limiter] struct {
	limiters             sync.Map
	replenishedPerSecond float64
	burst                int
	newLimiterFn         func(float64, int) Limiter
}

var _ Ratelimiter = &SyncMapLoadThenLoadOrStore[limiter.Limiter]{}

func NewSyncMapLoadThenLoadOrStore[Limiter limiter.Limiter](newLimiterFn func(float64, int) Limiter, replenishedPerSecond float64, burst int) *SyncMapLoadThenLoadOrStore[Limiter] {
	return &SyncMapLoadThenLoadOrStore[Limiter]{
		newLimiterFn:         newLimiterFn,
		replenishedPerSecond: replenishedPerSecond,
		burst:                burst,
	}
}

func (d *SyncMapLoadThenLoadOrStore[Limiter]) AllowN(key string, cost int) (bool, error) {
	l, ok := d.limiters.Load(key)
	// if key is not found, then create a new limiter and store it
	// reduces allocation by doing this
	if !ok {
		l, _ = d.limiters.LoadOrStore(key, d.newLimiterFn(d.replenishedPerSecond, d.burst))
	}
	return l.(limiter.Limiter).AllowN(cost), nil
}

func (d *SyncMapLoadThenLoadOrStore[Limiter]) ForceN(key string, cost int) (bool, error) {
	l, ok := d.limiters.Load(key)
	// if key is not found, then create a new limiter and store it
	// reduces allocation by doing this
	if !ok {
		l, _ = d.limiters.LoadOrStore(key, d.newLimiterFn(d.replenishedPerSecond, d.burst))
	}
	return l.(limiter.Limiter).AllowN(cost), nil
}

func (d *SyncMapLoadThenLoadOrStore[Limiter]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

type SyncMapLoadOrStore[Limiter limiter.Limiter] struct {
	limiters             sync.Map
	replenishedPerSecond float64
	burst                int
	constructor          func(float64, int) Limiter
}

func NewSyncMapLoadOrStore[Limiter limiter.Limiter](newLimiterFn func(float64, int) Limiter, replenishedPerSecond float64, burst int) *SyncMapLoadOrStore[Limiter] {
	return &SyncMapLoadOrStore[Limiter]{
		constructor:          newLimiterFn,
		replenishedPerSecond: replenishedPerSecond,
		burst:                burst,
	}
}

func (d *SyncMapLoadOrStore[Limiter]) AllowN(key string, cost int) (bool, error) {
	l, _ := d.limiters.LoadOrStore(key, d.constructor(d.replenishedPerSecond, d.burst))
	return l.(limiter.Limiter).AllowN(cost), nil
}

func (d *SyncMapLoadOrStore[Limiter]) ForceN(key string, n int) (bool, error) {
	l, _ := d.limiters.LoadOrStore(key, d.constructor(d.replenishedPerSecond, d.burst))
	return l.(limiter.Limiter).ForceN(n), nil
}

func (d *SyncMapLoadOrStore[Limiter]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

var _ Ratelimiter = &SyncMapLoadOrStore[limiter.Limiter]{}

type SyncMapLoadThenStore[Limiter limiter.Limiter] struct {
	limiters             sync.Map
	replenishedPerSecond float64
	burst                int
	newLimiterFn         func(float64, int) Limiter
}

var _ Ratelimiter = &SyncMapLoadThenStore[limiter.Limiter]{}

func NewSyncMapLoadThenStore[Limiter limiter.Limiter](newLimiterFn func(float64, int) Limiter, replenishPerSecond float64, burst int) *SyncMapLoadThenStore[Limiter] {
	return &SyncMapLoadThenStore[Limiter]{
		replenishedPerSecond: replenishPerSecond,
		burst:                burst,
		limiters:             sync.Map{},
		newLimiterFn:         newLimiterFn,
	}
}

func (r *SyncMapLoadThenStore[Limiter]) AllowN(key string, cost int) (bool, error) {
	l, ok := r.limiters.Load(key)
	if !ok {
		l = r.newLimiterFn(r.replenishedPerSecond, r.burst)
		// note: We might overwrite the limiter if multiple goroutines are trying to create the limiter at the same time, it is not a big deal as we may just have a little inaccurate rate limiting for a short period of time
		r.limiters.Store(key, l)
	}
	return l.(Limiter).AllowN(cost), nil
}

func (r *SyncMapLoadThenStore[Limiter]) GetLimiter(key string) Limiter {
	l, ok := r.limiters.Load(key)
	if !ok {
		l = limiter.NewResetbasedLimiter(r.replenishedPerSecond, r.burst)
		r.limiters.Store(key, l)
	}
	return l.(Limiter)
}

func (r *SyncMapLoadThenStore[Limiter]) Allow(key string) (bool, error) {
	return r.AllowN(key, 1)
}
