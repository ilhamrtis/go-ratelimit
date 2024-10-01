package ratelimit

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/yesyoukenspace/go-ratelimit/internal/utils"
)

func BenchmarkRatelimiters(b *testing.B) {
	rate := (100.0 / 60)
	burst := 100
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", MaxRetries: -1, DialTimeout: 20 * time.Millisecond, PoolTimeout: 20 * time.Millisecond, ContextTimeoutEnabled: true})

	ratelimiters := map[string]Ratelimiter{
		"SyncMap + LoadOrStore":        &SyncMapLoadOrStore{R: rate, B: burst},
		"SyncMap + Load > LoadOrStore": &SyncMapLoadThenLoadOrStore{R: rate, B: burst},
		"SyncMap + Load > Store":       &SyncMapLoadThenStore{R: rate, B: burst},
		"Map + Mutex":                  NewMutex(nil, rate, burst),
		"Map + RWMutex":                NewRWMutex(nil, rate, burst),
		"Redis":                        NewGoRedis(rdb, rate, burst),
	}

	for _, concurrentUsers := range []int{2048, 16384} {
		for name, limiter := range ratelimiters {
			ratelimitRunner(b, name, concurrentUsers, limiter)
		}
	}
}

func ratelimitRunner(b *testing.B, name string, numberOfKeys int, limiter Ratelimiter) bool {
	rStr := utils.RandString(4)
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
						if ok, _ := limiter.Allow(k); ok {
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
