package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type GoRedis struct {
	limiter *redis_rate.Limiter
	limit   redis_rate.Limit
	rdb     *redis.Client
}

var _ Ratelimiter = &GoRedis{}

func (d *GoRedis) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

func (d *GoRedis) AllowN(key string, n int) (bool, error) {
	// TODO: rate here only works for more than 1 rps, allow for less than 1 rps, and integers only
	res, err := d.limiter.AllowN(context.Background(), key, d.limit, n)
	if err != nil {
		return false, err
	}
	return res.Allowed > 0, nil
}

func NewGoRedis(rdb *redis.Client, rps float64, burst int) *GoRedis {

	return &GoRedis{
		limiter: redis_rate.NewLimiter(rdb),
		limit:   redis_rate.Limit{Rate: int(rps), Burst: burst, Period: time.Second},
	}
}
