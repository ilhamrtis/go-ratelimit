package ratelimit

import (
	"sync"
)

type LiGrMutex struct {
	mu sync.Mutex
	M  map[string]Limiter
	R  ReqPerSec
	B  int
}

var _ LimiterGroup = &LiGrMutex{}

func NewLiGrMutex(reqPerSec ReqPerSec, burst int) *LiGrMutex {
	return &LiGrMutex{
		M: make(map[string]Limiter),
		R: reqPerSec,
		B: burst,
	}
}

func (d *LiGrMutex) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	l, ok := d.M[key]
	if !ok {
		l = NewDefaultLimiter(d.R, d.B)
		d.M[key] = l
	}
	return l.Allow()
}

type LiGrRWMutex struct {
	mu sync.RWMutex
	M  map[string]Limiter
	R  ReqPerSec
	B  int
}

func NewLiGrRWMutex(reqPerSec ReqPerSec, burst int) *LiGrRWMutex {
	return &LiGrRWMutex{
		M: make(map[string]Limiter),
		R: reqPerSec,
		B: burst,
	}
}

func (d *LiGrRWMutex) Allow(key string) bool {
	d.mu.RLock()
	l, ok := d.M[key]
	if !ok {
		d.mu.RUnlock()
		d.mu.Lock()
		defer d.mu.Unlock()
		l = NewDefaultLimiter(d.R, d.B)
		d.M[key] = l
	} else {
		d.mu.RUnlock()
	}
	return l.Allow()
}
