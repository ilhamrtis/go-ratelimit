package limiter

type Limiter interface {
	Allow() bool
	AllowN(int) bool
	ForceN(int) bool
}
