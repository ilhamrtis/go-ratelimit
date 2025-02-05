package limiter

import (
	"math"
	"sync"
	"time"
)

type ResetBasedLimiter struct {
	resetAt    float64
	B          int
	Rps        float64
	Rpms       float64
	mu         sync.Mutex
	burstInMs  float64
	deltaSince float64
}

func NewResetbasedLimiter(reqPerSec float64, burst int) *ResetBasedLimiter {
	rpms := reqPerSec / 1000
	burstInMs := float64(burst) / rpms
	return &ResetBasedLimiter{
		Rps:       reqPerSec,
		Rpms:      rpms,
		B:         burst,
		burstInMs: burstInMs,
		resetAt:   float64(time.Now().UnixMilli()) - burstInMs,
	}
}

func (l *ResetBasedLimiter) Allow() (bool, error) {
	return l.AllowN(1)
}

func (l *ResetBasedLimiter) allowN(n int, shouldCheck bool) (bool, error) {
	now := float64(time.Now().UnixMilli())
	if (shouldCheck && l.resetAt > now) || n > l.B {
		return false, nil
	}

	incrementInMs := float64(n) / l.Rpms
	l.mu.Lock()
	defer l.mu.Unlock()
	newResetAt := math.Max(l.resetAt, now-l.burstInMs) + incrementInMs
	if shouldCheck && newResetAt > now {
		return false, nil
	}
	l.resetAt = newResetAt
	l.deltaSince += incrementInMs
	return true, nil
}

func (l *ResetBasedLimiter) AllowN(n int) (bool, error) {
	return l.allowN(n, true)
}

func (l *ResetBasedLimiter) ForceN(n int) (bool, error) {
	return l.allowN(n, false)
}

func (l *ResetBasedLimiter) IncrementResetAtBy(ms float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.resetAt += ms
}

func (l *ResetBasedLimiter) GetResetAt() float64 {
	return l.resetAt
}

func (l *ResetBasedLimiter) PopDeltaSinceInMs() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	deltaSince := l.deltaSince
	l.deltaSince = 0
	return deltaSince
}
