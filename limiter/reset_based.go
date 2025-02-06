package limiter

import (
	"math"
	"sync"
	"time"
)

type ResetBasedLimiter struct {
	resetAt        float64
	burst          int
	tokenPerSecond float64
	tokenPerNano   float64
	mu             sync.Mutex
	burstInNano    float64
	deltaSince     float64
}

func NewResetbasedLimiter(tps float64, burst int) *ResetBasedLimiter {
	tpns := tps / float64(time.Second/time.Nanosecond)
	burstInNano := float64(burst) / tpns
	return &ResetBasedLimiter{
		tokenPerSecond: tps,
		tokenPerNano:   tpns,
		burst:          burst,
		burstInNano:    burstInNano,
		resetAt:        float64(time.Now().UnixMilli()) - burstInNano,
	}
}

func (l *ResetBasedLimiter) Allow() (bool, error) {
	return l.AllowN(1)
}

func (l *ResetBasedLimiter) allowN(n int, shouldCheck bool) (bool, error) {
	now := float64(time.Now().UnixNano())
	if (shouldCheck && l.resetAt > now) || n > l.burst {
		return false, nil
	}

	incrementInNano := float64(n) / l.tokenPerNano
	l.mu.Lock()
	defer l.mu.Unlock()
	newResetAt := math.Max(l.resetAt, now-l.burstInNano) + incrementInNano
	if shouldCheck && newResetAt > now {
		return false, nil
	}
	l.resetAt = newResetAt
	l.deltaSince += incrementInNano
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
