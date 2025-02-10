package ratelimit

import (
	"sync"

	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type Mutex[L limiter.Limiter] struct {
	mu sync.Mutex
	M  map[string]L
	R  float64
	B  int
	c  func(float64, int) L
}

var _ Ratelimiter = &Mutex[limiter.Limiter]{}

func NewMutex[L limiter.Limiter](constuctor func(float64, int) L, reqPerSec float64, burst int) *Mutex[L] {
	return &Mutex[L]{
		M: make(map[string]L),
		R: reqPerSec,
		B: burst,
		c: constuctor,
	}
}

func (d *Mutex[L]) getLimiter(key string) L {
	d.mu.Lock()
	defer d.mu.Unlock()
	l, ok := d.M[key]
	if !ok {
		l = d.c(d.R, d.B)
		d.M[key] = l
	}
	return l
}

func (d *Mutex[L]) AllowN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.AllowN(n), nil
}

func (d *Mutex[L]) ForceN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.ForceN(n), nil
}

func (d *Mutex[L]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

type RWMutex[L limiter.Limiter] struct {
	mu sync.RWMutex
	M  map[string]L
	R  float64
	B  int
	c  func(float64, int) L
}

func NewRWMutex[L limiter.Limiter](constructor func(float64, int) L, reqPerSec float64, burst int) *RWMutex[L] {
	return &RWMutex[L]{
		M: make(map[string]L),
		R: reqPerSec,
		B: burst,
		c: constructor,
	}
}

func (d *RWMutex[L]) getLimiter(key string) limiter.Limiter {
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

func (d *RWMutex[L]) AllowN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.AllowN(n), nil
}

func (d *RWMutex[L]) ForceN(key string, n int) (bool, error) {
	l := d.getLimiter(key)
	return l.ForceN(n), nil
}

func (d *RWMutex[L]) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}
