package ratelimit

import (
	"sync"

	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type SyncMapLoadThenLoadOrStore[L limiter.Limiter] struct {
	m sync.Map
	R float64
	B int
	c func(float64, int) L
}

var _ Ratelimiter = &SyncMapLoadThenLoadOrStore[limiter.Limiter]{}

func NewSyncMapLoadThenLoadOrStore[L limiter.Limiter](constructor func(float64, int) L, reqPerSec float64, burst int) *SyncMapLoadThenLoadOrStore[L] {
	return &SyncMapLoadThenLoadOrStore[L]{
		c: constructor,
		R: reqPerSec,
		B: burst,
	}
}

func (d *SyncMapLoadThenLoadOrStore[L]) AllowN(key string, n int) (bool, error) {
	l, ok := d.m.Load(key)
	// if key is not found, then create a new limiter and store it
	// reduces allocation by doing this
	if !ok {
		l, _ = d.m.LoadOrStore(key, d.c(d.R, d.B))
	}
	return l.(limiter.Limiter).AllowN(n), nil
}

func (d *SyncMapLoadThenLoadOrStore[L]) ForceN(key string, n int) (bool, error) {
	l, ok := d.m.Load(key)
	// if key is not found, then create a new limiter and store it
	// reduces allocation by doing this
	if !ok {
		l, _ = d.m.LoadOrStore(key, d.c(d.R, d.B))
	}
	return l.(limiter.Limiter).AllowN(n), nil
}

func (d *SyncMapLoadThenLoadOrStore[L]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

type SyncMapLoadOrStore[L limiter.Limiter] struct {
	m sync.Map
	R float64
	B int
	c func(float64, int) L
}

func NewSyncMapLoadOrStore[L limiter.Limiter](constructor func(float64, int) L, reqPerSec float64, burst int) *SyncMapLoadOrStore[L] {
	return &SyncMapLoadOrStore[L]{
		c: constructor,
		R: reqPerSec,
		B: burst,
	}
}

func (d *SyncMapLoadOrStore[L]) AllowN(key string, n int) (bool, error) {
	l, _ := d.m.LoadOrStore(key, d.c(d.R, d.B))
	return l.(limiter.Limiter).AllowN(n), nil
}

func (d *SyncMapLoadOrStore[L]) ForceN(key string, n int) (bool, error) {
	l, _ := d.m.LoadOrStore(key, d.c(d.R, d.B))
	return l.(limiter.Limiter).ForceN(n), nil
}

func (d *SyncMapLoadOrStore[L]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

var _ Ratelimiter = &SyncMapLoadOrStore[limiter.Limiter]{}

type SyncMapLoadThenStore[L limiter.Limiter] struct {
	m     sync.Map
	tps   float64
	burst int
	c     func(float64, int) L
}

var _ Ratelimiter = &SyncMapLoadThenStore[limiter.Limiter]{}

func NewSyncMapLoadThenStore[L limiter.Limiter](constructor func(float64, int) L, rps float64, burst int) *SyncMapLoadThenStore[L] {
	return &SyncMapLoadThenStore[L]{
		tps:   rps,
		burst: burst,
		m:     sync.Map{},
		c:     constructor,
	}
}

func (r *SyncMapLoadThenStore[L]) AllowN(key string, n int) (bool, error) {
	l, ok := r.m.Load(key)
	if !ok {
		l = r.c(r.tps, r.burst)
		// note: We might overwrite the limiter if multiple goroutines are trying to create the limiter at the same time, it is not a big deal as we may just have a little inaccurate rate limiting for a short period of time
		r.m.Store(key, l)
	}
	return l.(L).AllowN(n), nil
}

func (r *SyncMapLoadThenStore[L]) GetLimiter(key string) L {
	l, ok := r.m.Load(key)
	if !ok {
		l = limiter.NewResetbasedLimiter(r.tps, r.burst)
		r.m.Store(key, l)
	}
	return l.(L)
}

func (r *SyncMapLoadThenStore[l]) Allow(key string) (bool, error) {
	return r.AllowN(key, 1)
}
