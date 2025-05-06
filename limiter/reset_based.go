package limiter

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type ResetBasedLimiter struct {
	resetAt            atomic.Int64
	burst              int
	nanosecondPerToken int64
	mu                 sync.RWMutex
	burstInNano        int64
	deltaSince         atomic.Int64
}

var _ Limiter = &ResetBasedLimiter{}

func NewResetbasedLimiter(tps float64, burst int) *ResetBasedLimiter {
	nspt := int64(float64(time.Second) / tps)
	burstInNano := int64(burst) * nspt
	l := &ResetBasedLimiter{
		nanosecondPerToken: nspt,
		burst:              burst,
		burstInNano:        burstInNano,
	}
	l.resetAt.Store(time.Now().UnixNano() - burstInNano)
	return l
}

func (l *ResetBasedLimiter) Allow() bool {
	return l.AllowN(1)
}

func (l *ResetBasedLimiter) allowN(n int, shouldCheck bool) bool {
	now := time.Now().UnixNano()
	if (shouldCheck && l.resetAt.Load() > now) || n > l.burst {
		return false
	}

	incrementInNano := int64(n) * l.nanosecondPerToken
	l.mu.Lock()
	defer l.mu.Unlock()
	newResetAt := math.Max(float64(l.resetAt.Load()), float64(now-l.burstInNano)) + float64(incrementInNano)
	if shouldCheck && newResetAt > float64(now) {
		return false
	}
	l.resetAt.Store(int64(newResetAt))
	l.deltaSince.Add(incrementInNano)
	return true
}

func (l *ResetBasedLimiter) AllowN(n int) bool {
	return l.allowN(n, true)
}

func (l *ResetBasedLimiter) ForceN(n int) bool {
	return l.allowN(n, false)
}

func (l *ResetBasedLimiter) IncrementResetAtBy(inc int64) {
	l.resetAt.Add(inc)
}

func (l *ResetBasedLimiter) GetResetAt() int64 {
	return l.resetAt.Load()
}

func (l *ResetBasedLimiter) PopResetAtDelta() int64 {
	return l.deltaSince.Swap(0)
}
