package ratelimit

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type GoRedisRate struct {
	ctx     context.Context
	limiter *redis_rate.Limiter
	limit   redis_rate.Limit
}

var _ Ratelimiter = &GoRedisRate{}

func (d *GoRedisRate) Allow(key string) (bool, error) {
	return d.AllowN(key, 1)
}

func (d *GoRedisRate) AllowN(key string, cost int) (bool, error) {
	res, err := d.limiter.AllowN(d.ctx, key, d.limit, cost)
	if err != nil {
		return false, err
	}
	return res.Allowed > 0, nil
}

// ForceN is not implemented for GoRedisRate
func (d *GoRedisRate) ForceN(key string, cost int) (bool, error) {
	return d.AllowN(key, cost)
}

func NewGoRedis(redisClient *redis.Client, replenishedPerSecond float64, burst int) *GoRedisRate {
	return &GoRedisRate{
		ctx:     context.Background(),
		limiter: redis_rate.NewLimiter(redisClient),
		// TODO: rate here only works for more than 1 rps, allow for less than 1 rps, and integers only
		limit: redis_rate.Limit{Rate: int(replenishedPerSecond), Burst: burst, Period: time.Second},
	}
}
