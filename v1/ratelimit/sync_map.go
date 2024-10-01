package ratelimit

import (
	"sync"

	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type SyncMapLoadThenLoadOrStore struct {
	m sync.Map
	R float64
	B int
	c NewLimiterFn
}

var _ Ratelimiter = &SyncMapLoadThenLoadOrStore{}

func NewSyncMapLoadThenLoadOrStore(constructor NewLimiterFn, reqPerSec float64, burst int) *SyncMapLoadThenLoadOrStore {
	if constructor == nil {
		constructor = NewDefaultLimiter
	}
	return &SyncMapLoadThenLoadOrStore{
		c: constructor,
		R: reqPerSec,
		B: burst,
	}
}

func (d *SyncMapLoadThenLoadOrStore) Allow(key string) (bool, error) {
	l, ok := d.m.Load(key)
	// if key is not found, then create a new limiter and store it
	// reduces allocation by doing this
	if !ok {
		l, _ = d.m.LoadOrStore(key, d.c(d.R, d.B))
	}
	return l.(limiter.Limiter).Allow()
}

type SyncMapLoadOrStore struct {
	m sync.Map
	R float64
	B int
	c NewLimiterFn
}

func NewSyncMapLoadOrStore(constructor NewLimiterFn, reqPerSec float64, burst int) *SyncMapLoadOrStore {
	if constructor == nil {
		constructor = NewDefaultLimiter
	}

	return &SyncMapLoadOrStore{
		c: constructor,
		R: reqPerSec,
		B: burst,
	}
}

func (d *SyncMapLoadOrStore) Allow(key string) (bool, error) {
	l, _ := d.m.LoadOrStore(key, d.c(d.R, d.B))
	return l.(limiter.Limiter).Allow()
}

var _ Ratelimiter = &SyncMapLoadOrStore{}

type SyncMapLoadThenStore struct {
	m sync.Map
	R float64
	B int
	c NewLimiterFn
}

func NewSyncMapLoadThenStore(constructor NewLimiterFn, reqPerSec float64, burst int) *SyncMapLoadThenStore {
	if constructor == nil {
		constructor = NewDefaultLimiter
	}
	return &SyncMapLoadThenStore{
		c: constructor,
		R: reqPerSec,
		B: burst,
	}
}

func (d *SyncMapLoadThenStore) Allow(key string) (bool, error) {
	l, ok := d.m.Load(key)
	if !ok {
		l = NewDefaultLimiter(d.R, d.B)
		// note: We might overwrite the limiter if multiple goroutines are trying to create the limiter at the same time, it is not a big deal as we may just have a little inaccurate rate limiting for a short period of time
		d.m.Store(key, l)
	}
	return l.(limiter.Limiter).Allow()
}
