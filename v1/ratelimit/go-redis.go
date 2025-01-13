package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type GoRedisRate struct {
	limiter *redis_rate.Limiter
	limit   redis_rate.Limit
	rdb     *redis.Client
}

var _ Ratelimiter = &GoRedisRate{}

func (d *GoRedisRate) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

func (d *GoRedisRate) AllowN(key string, n int) (bool, error) {
	res, err := d.limiter.AllowN(context.Background(), key, d.limit, n)
	if err != nil {
		return false, err
	}
	return res.Allowed > 0, nil
}

func NewGoRedis(rdb *redis.Client, rps float64, burst int) *GoRedisRate {
	return &GoRedisRate{
		limiter: redis_rate.NewLimiter(rdb),
		// TODO: rate here only works for more than 1 rps, allow for less than 1 rps, and integers only
		limit: redis_rate.Limit{Rate: int(rps), Burst: burst, Period: time.Second},
	}
}
