package ratelimit_test

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/yesyoukenspace/ratelimit"
)

func BenchmarkLimiterGroup(b *testing.B) {
	rate := ratelimit.ReqPerSec(100 / 60)
	burst := 100
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", MaxRetries: -1, DialTimeout: 20 * time.Millisecond, PoolTimeout: 20 * time.Millisecond, ContextTimeoutEnabled: true})

	limiterGroups := map[string]ratelimit.LimiterGroup{
		"SyncMap + LoadOrStore":        &ratelimit.LiGrSyncMapLoadOrStore{R: rate, B: burst},
		"SyncMap + Load > LoadOrStore": &ratelimit.LiGrSyncMapLoadThenLoadOrStore{R: rate, B: burst},
		"SyncMap + Load > Store":       &ratelimit.LiGrSyncMapLoadThenStore{R: rate, B: burst},
		"Map + Mutex":                  ratelimit.NewLiGrMutex(rate, burst),
		"Map + RWMutex":                ratelimit.NewLiGrRWMutex(rate, burst),
		"Redis":                        ratelimit.NewLiGrRedis(rdb, rate, burst),
	}

	for _, concurrentUsers := range []int{2048, 16384} {
		for name, limiter := range limiterGroups {
			limiterGroupRunner(b, name, concurrentUsers, limiter)
		}
	}
}

func limiterGroupRunner(b *testing.B, name string, numberOfKeys int, limiter ratelimit.LimiterGroup) bool {
	rStr := randString(4)
	return b.Run(fmt.Sprintf("name:%s;number of keys:%d", name, numberOfKeys), func(b *testing.B) {
		b.ReportAllocs()

		n := b.N
		batchSize := n / numberOfKeys
		var wgg sync.WaitGroup
		if batchSize == 0 {
			batchSize = n
		}
		key := 0
		mu := sync.RWMutex{}
		mu.Lock()
		allowed := 0
		disallowed := 0

		for n > 0 {
			wgg.Add(1)

			batch := min(n, batchSize)
			n -= batch
			go func(k string, quota int) {
				var wg sync.WaitGroup
				wg.Add(quota)
				mu.RLock()
				defer mu.RUnlock()
				for i := 0; i < quota; i++ {
					go func() {
						if limiter.Allow(k) {
							allowed++
						} else {
							disallowed++
						}
						wg.Done()
					}()
				}
				wg.Wait()
				wgg.Done()
			}(fmt.Sprintf("%s#%s", strconv.Itoa(key), rStr), batch)
			key++
		}
		b.StartTimer()
		mu.Unlock()
		wgg.Wait()
		b.StopTimer()
		runtime.KeepAlive(allowed)
		runtime.KeepAlive(disallowed)
	})
}
