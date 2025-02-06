package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type RedisDelayedSync struct {
	syncInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	inner        *SyncMapLoadThenStore[*limiter.ResetBasedLimiter]
	rdb          *redis.Client
	currentIndex atomic.Int32
	toSync       []sync.Map
	lastSynced   map[string]int64
	opt          RedisDelayedSyncOption
	burstInMs    float64
}

type RedisDelayedSyncOption struct {
	// SyncInterval is the interval to sync the rate limit to the redis
	// Adjust this value to trade off between the performance and the accuracy of the rate limit
	SyncInterval time.Duration
	TokenPerSec  float64
	Burst        int
	RedisClient  *redis.Client
}

func NewRedisDelayedSync(ctx context.Context, opt RedisDelayedSyncOption) *RedisDelayedSync {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	rl := &RedisDelayedSync{
		ctx:          ctx,
		cancel:       cancel,
		rdb:          opt.RedisClient,
		syncInterval: opt.SyncInterval,
		inner:        NewSyncMapLoadThenStore(limiter.NewResetbasedLimiter, opt.TokenPerSec, opt.Burst),
		toSync:       []sync.Map{{}, {}},
		lastSynced:   map[string]int64{},
		opt:          opt,
		burstInMs:    float64(opt.Burst) / opt.TokenPerSec * 1000,
	}
	go func() {
		ticker := time.NewTicker(opt.SyncInterval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				indexToSync := rl.currentIndex.Load()
				rl.currentIndex.Store((indexToSync + 1) % int32(len(rl.toSync)))
				if err := rl.syncAll(indexToSync); err != nil {
					// TODO: handle error
					fmt.Printf("error syncing: %v\n", err)
				}
			}
		}
	}()
	return rl
}

func (r *RedisDelayedSync) Allow(key string) (bool, error) {
	return r.AllowN(key, 1)
}

func (r *RedisDelayedSync) AllowN(key string, n int) (bool, error) {
	ok, err := r.inner.AllowN(key, n)
	if ok {
		r.toSync[r.currentIndex.Load()].LoadOrStore(key, struct{}{})
	}
	return ok, err
}

func (r *RedisDelayedSync) syncAll(index int32) error {
	r.toSync[index].Range(func(key, value any) bool {
		if err := r.sync(key.(string)); err != nil {
			return false
		}
		return true
	})
	go r.toSync[index].Clear()
	return nil
}

func (r *RedisDelayedSync) sync(key string) error {
	limiter := r.inner.GetLimiter(key)
	incrementInMs := limiter.PopDeltaSinceInMs()
	if r.lastSynced[key] == 0 {
		resetAt := limiter.GetResetAt()
		cmd := r.rdb.SetNX(r.ctx, key, int64(resetAt), time.Hour*48)
		if cmd.Err() != nil {
			return cmd.Err()
		}
		r.lastSynced[key] = int64(resetAt)
	}
	cmd := r.rdb.IncrBy(r.ctx, key, int64(incrementInMs))
	if cmd.Err() != nil {
		return cmd.Err()
	}

	sharedResetAt := float64(cmd.Val())
	diff := sharedResetAt - float64(r.lastSynced[key]) - incrementInMs
	if diff > 0 {
		limiter.IncrementResetAtBy(float64(diff))
	}
	r.lastSynced[key] = int64(sharedResetAt)
	return nil
}
