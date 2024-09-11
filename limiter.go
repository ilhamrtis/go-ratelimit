package ratelimit

import (
	"sync"
	"time"
)

type Bucket struct {
	B         float64
	R         ReqPerSec
	remaining float64
	lastCheck time.Time
	mu        sync.Mutex
}

func NewBucket(rate ReqPerSec, burst int) *Bucket {
	return &Bucket{
		B:         float64(burst),
		remaining: float64(burst),
		R:         rate,
		lastCheck: time.Now(),
	}
}

func (b *Bucket) AllowN(n int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	leak := float64(now.Second()-b.lastCheck.Second()) * b.R

	if leak > 0 {
		if b.remaining+leak > b.B {
			b.remaining = b.B
		} else {
			b.remaining += leak
		}
		b.lastCheck = now
	}
	nFloat := float64(n)
	if b.remaining >= nFloat {
		b.remaining -= nFloat
		return true
	} else {
		return false
	}
}

func (b *Bucket) Allow() bool {
	return b.AllowN(1)
}
