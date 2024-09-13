package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type LiGrRedis struct {
	m     sync.Map
	limit redis_rate.Limit
	rdb   *redis.Client
}

var _ LimiterGroup = &LiGrRedis{}

func (d *LiGrRedis) Allow(key string) bool {
	return d.AllowN(key, 1)
}

func (d *LiGrRedis) AllowN(key string, n int) bool {
	l, ok := d.m.Load(key)
	// if key is not found, then create a new limiter and store it
	// reduces allocation by doing this
	if !ok {
		l, _ = d.m.LoadOrStore(key, redis_rate.NewLimiter(d.rdb))
	}

	res, err := l.(*redis_rate.Limiter).AllowN(context.Background(), key, d.limit, n)
	if err != nil {
		// TODO: reconsider this error handling
		return false
	}
	return res.Allowed > 0
}

func NewLiGrRedis(rdb *redis.Client, rps ReqPerSec, burst int) *LiGrRedis {
	return &LiGrRedis{
		rdb: rdb,
		// TODO: rate here only works for more than 1 rps, allow for less than 1 rps
		limit: redis_rate.Limit{Rate: int(rps), Burst: burst, Period: time.Second},
	}
}
