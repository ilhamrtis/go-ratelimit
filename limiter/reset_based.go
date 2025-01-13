package limiter

import (
	"sync"
	"time"
)

type ResetBasedLimiter struct {
	resetAt float64
	B       int
	Rps     float64
	mu      sync.Mutex
}

func NewResetbasedLimiter(reqPerSec float64, burst int) *ResetBasedLimiter {
	return &ResetBasedLimiter{
		Rps:     reqPerSec,
		B:       burst,
		resetAt: float64(time.Now().Second()) - float64(burst)/reqPerSec,
	}
}

func (l *ResetBasedLimiter) Allow() (bool, error) {
	return l.AllowN(1)
}

func (l *ResetBasedLimiter) AllowN(n int) (bool, error) {

	now := float64(time.Now().Second())
	if l.resetAt > now || n > l.B {
		return false, nil
	}

	increment := float64(n) / l.Rps
	l.mu.Lock()
	defer l.mu.Unlock()
	newResetAt := l.resetAt + increment

	if newResetAt > now {
		return false, nil
	}
	l.resetAt = newResetAt
	return true, nil
}
