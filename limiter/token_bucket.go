package limiter

import (
	"sync"
	"time"
)

type Bucket struct {
	B         float64
	R         float64
	remaining float64
	lastCheck time.Time
	mu        sync.Mutex
}

func NewBucket(rate float64, burst int) *Bucket {
	return &Bucket{
		B:         float64(burst),
		remaining: float64(burst),
		R:         rate,
		lastCheck: time.Now(),
	}
}

func (b *Bucket) ForceN(n int) bool {
	return b.allowN(n, false)
}

func (b *Bucket) AllowN(n int) bool {
	return b.allowN(n, true)
}

func (b *Bucket) allowN(n int, shouldCheck bool) bool {
	now := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()
	leak := now.Sub(b.lastCheck).Seconds() * b.R

	if leak > 0 {
		if b.remaining+leak > b.B {
			b.remaining = b.B
		} else {
			b.remaining += leak
		}
		b.lastCheck = now
	}
	nFloat := float64(n)
	if !shouldCheck || b.remaining >= nFloat {
		b.remaining -= nFloat
		return true
	} else {
		return false
	}
}

func (b *Bucket) Allow() bool {
	return b.AllowN(1)
}
