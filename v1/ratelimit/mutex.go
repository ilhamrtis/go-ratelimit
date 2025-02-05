package ratelimit

import (
	"sync"

	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type Mutex struct {
	mu sync.Mutex
	M  map[string]limiter.Limiter
	R  float64
	B  int
	c  NewLimiterFn
}

var _ Ratelimiter = &Mutex{}

func NewMutex(constuctor NewLimiterFn, reqPerSec float64, burst int) *Mutex {
	if constuctor == nil {
		constuctor = NewDefaultLimiter
	}
	return &Mutex{
		M: make(map[string]limiter.Limiter),
		R: reqPerSec,
		B: burst,
		c: constuctor,
	}
}

func (d *Mutex) getLimiter(key string) limiter.Limiter {
	d.mu.Lock()
	defer d.mu.Unlock()
	l, ok := d.M[key]
	if !ok {
		l = d.c(d.R, d.B)
		d.M[key] = l
	}
	return l
}

func (d *Mutex) AllowN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.AllowN(n)
}

func (d *Mutex) ForceN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.ForceN(n)
}

func (d *Mutex) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

type RWMutex struct {
	mu sync.RWMutex
	M  map[string]limiter.Limiter
	R  float64
	B  int
	c  NewLimiterFn
}

func NewRWMutex(constructor NewLimiterFn, reqPerSec float64, burst int) *RWMutex {
	if constructor == nil {
		constructor = NewDefaultLimiter
	}
	return &RWMutex{
		M: make(map[string]limiter.Limiter),
		R: reqPerSec,
		B: burst,
		c: constructor,
	}
}

func (d *RWMutex) getLimiter(key string) limiter.Limiter {
	d.mu.RLock()
	l, ok := d.M[key]
	d.mu.RUnlock()
	if !ok {
		d.mu.Lock()
		defer d.mu.Unlock()
		l = d.c(d.R, d.B)
		d.M[key] = l
	}
	return l
}

func (d *RWMutex) AllowN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.AllowN(n)
}

func (d *RWMutex) ForceN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.ForceN(n)
}

func (d *RWMutex) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}
