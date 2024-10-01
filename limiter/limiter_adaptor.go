package limiter

type SimpleLimiter interface {
	Allow() bool
}

type SimpleLimiterAdapter struct {
	SimpleLimiter
}

func (s SimpleLimiterAdapter) Allow() (bool, error) {
	return s.SimpleLimiter.Allow(), nil
}

func NewSimpleLimiterAdapter(l SimpleLimiter) *SimpleLimiterAdapter {
	return &SimpleLimiterAdapter{l}
}
