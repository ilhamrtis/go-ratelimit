package ratelimit

import (
	"sync"

	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type Mutex[Limiter limiter.Limiter] struct {
	mu                   sync.Mutex
	limiters             map[string]Limiter
	replenishedPerSecond float64
	burst                int
	newLimiterFn         func(float64, int) Limiter
}

var _ Ratelimiter = &Mutex[limiter.Limiter]{}

func NewMutex[Limiter limiter.Limiter](newLimiterFn func(float64, int) Limiter, replenishedPerSecond float64, burst int) *Mutex[Limiter] {
	return &Mutex[Limiter]{
		limiters:             make(map[string]Limiter),
		replenishedPerSecond: replenishedPerSecond,
		burst:                burst,
		newLimiterFn:         newLimiterFn,
	}
}

func (d *Mutex[Limiter]) getLimiter(key string) Limiter {
	d.mu.Lock()
	defer d.mu.Unlock()
	l, ok := d.limiters[key]
	if !ok {
		l = d.newLimiterFn(d.replenishedPerSecond, d.burst)
		d.limiters[key] = l
	}
	return l
}

func (d *Mutex[Limiter]) AllowN(key string, cost int) (bool, error) {
	l := d.getLimiter(key)
	return l.AllowN(cost), nil
}

func (d *Mutex[Limiter]) ForceN(key string, cost int) (bool, error) {
	l := d.getLimiter(key)
	return l.ForceN(cost), nil
}

func (d *Mutex[Limiter]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

type RWMutex[Limiter limiter.Limiter] struct {
	mu                   sync.RWMutex
	limiters             map[string]Limiter
	replenishedPerSecond float64
	burst                int
	newLimiterFn         func(float64, int) Limiter
}

func NewRWMutex[Limiter limiter.Limiter](newLimiterFn func(float64, int) Limiter, replenishedPerSecond float64, burst int) *RWMutex[Limiter] {
	return &RWMutex[Limiter]{
		limiters:             make(map[string]Limiter),
		replenishedPerSecond: replenishedPerSecond,
		burst:                burst,
		newLimiterFn:         newLimiterFn,
	}
}

func (d *RWMutex[Limiter]) getLimiter(key string) limiter.Limiter {
	d.mu.RLock()
	l, ok := d.limiters[key]
	d.mu.RUnlock()
	if !ok {
		d.mu.Lock()
		defer d.mu.Unlock()
		l = d.newLimiterFn(d.replenishedPerSecond, d.burst)
		d.limiters[key] = l
	}
	return l
}

func (d *RWMutex[Limiter]) AllowN(key string, cost int) (bool, error) {
	l := d.getLimiter(key)
	return l.AllowN(cost), nil
}

func (d *RWMutex[Limiter]) ForceN(key string, cost int) (bool, error) {
	l := d.getLimiter(key)
	return l.ForceN(cost), nil
}

func (d *RWMutex[Limiter]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}
