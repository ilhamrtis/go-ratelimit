package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yesyoukenspace/go-ratelimit/limiter"
)

type rl1[l limiter.Limiter] struct {
	m     sync.Map
	rps   float64
	burst int
}

func NewRl1[L limiter.Limiter](rps float64, burst int) *rl1[L] {
	return &rl1[L]{
		rps:   rps,
		burst: burst,
		m:     sync.Map{},
	}
}

func (r *rl1[L]) AllowN(key string, n int) (bool, error) {
	l, ok := r.m.Load(key)
	if !ok {
		l = limiter.NewResetbasedLimiter(r.rps, r.burst)
		r.m.Store(key, l)

	}
	return l.(L).AllowN(n)
}

func (r *rl1[L]) GetLimiter(key string) L {
	l, ok := r.m.Load(key)
	if !ok {
		l = limiter.NewResetbasedLimiter(r.rps, r.burst)
		r.m.Store(key, l)
	}
	return l.(L)
}

func (r *rl1[l]) Allow(key string) (bool, error) {
	return r.AllowN(key, 1)
}

type RedisWithDelay struct {
	syncInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	inner        *rl1[*limiter.ResetBasedLimiter]
	rdb          *redis.Client
	currentIndex int
	toSync       []map[string]struct{}
	lastSynced   map[string]int64
	opt          RedisWithDelayOption
	burstInMs    float64
}

type RedisWithDelayOption struct {
	SyncInterval  time.Duration
	RequestPerSec float64
	Burst         int

	RedisClient *redis.Client
}

func NewRedisWithDelay(ctx context.Context, opt RedisWithDelayOption) *RedisWithDelay {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	rl := &RedisWithDelay{
		ctx:          ctx,
		cancel:       cancel,
		rdb:          opt.RedisClient,
		syncInterval: opt.SyncInterval,
		inner:        NewRl1[*limiter.ResetBasedLimiter](opt.RequestPerSec, opt.Burst),
		toSync:       []map[string]struct{}{{}, {}},
		lastSynced:   map[string]int64{},
		opt:          opt,
		burstInMs:    float64(opt.Burst) / opt.RequestPerSec * 1000,
	}
	go func() {
		ticker := time.NewTicker(opt.SyncInterval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				indexToSync := rl.currentIndex
				nextIndex := (rl.currentIndex + 1) % len(rl.toSync)
				rl.toSync[nextIndex] = map[string]struct{}{}
				rl.currentIndex = nextIndex
				if err := rl.syncAll(indexToSync); err != nil {
					// TODO: handle error
					fmt.Printf("error syncing: %v\n", err)
				}
			}
		}
	}()
	return rl
}

func (r *RedisWithDelay) Allow(key string) (bool, error) {
	return r.AllowN(key, 1)
}

func (r *RedisWithDelay) AllowN(key string, n int) (bool, error) {
	ok, err := r.inner.AllowN(key, n)
	if ok {
		r.toSync[r.currentIndex][key] = struct{}{}
	}
	return ok, err
}

func (r *RedisWithDelay) syncAll(index int) error {
	for key, _ := range r.toSync[index] {
		if err := r.sync(key); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisWithDelay) sync(key string) error {
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
