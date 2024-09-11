package ratelimit

import (
	"sync"
)

type LiGrSyncMapLoadThenLoadOrStore struct {
	m sync.Map
	R ReqPerSec
	B int
}

var _ LimiterGroup = &LiGrSyncMapLoadThenLoadOrStore{}

func (d *LiGrSyncMapLoadThenLoadOrStore) Allow(key string) bool {
	l, ok := d.m.Load(key)
	// if key is not found, then create a new limiter and store it
	// reduces allocation by doing this
	if !ok {
		l, _ = d.m.LoadOrStore(key, NewDefaultLimiter(d.R, d.B))
	}
	return l.(Limiter).Allow()
}

type LiGrSyncMapLoadOrStore struct {
	m sync.Map
	R ReqPerSec
	B int
}

func (d *LiGrSyncMapLoadOrStore) Allow(key string) bool {
	l, _ := d.m.LoadOrStore(key, NewDefaultLimiter(d.R, d.B))
	return l.(Limiter).Allow()
}

var _ LimiterGroup = &LiGrSyncMapLoadOrStore{}

type LiGrSyncMapLoadThenStore struct {
	m sync.Map
	R ReqPerSec
	B int
}

func (d *LiGrSyncMapLoadThenStore) Allow(key string) bool {
	l, ok := d.m.Load(key)
	if !ok {
		l = NewDefaultLimiter(d.R, d.B)
		// note: We might overwrite the limiter if multiple goroutines are trying to create the limiter at the same time, it is not a big deal as we may just have a little inaccurate rate limiting for a short period of time
		d.m.Store(key, l)
	}
	return l.(Limiter).Allow()
}
