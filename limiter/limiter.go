package limiter

type Limiter interface {
	Allow() (bool, error)
	AllowN(int) (bool, error)
	ForceN(int) (bool, error)
}
