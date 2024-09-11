package ratelimit_test

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"

	"testing"

	"github.com/yesyoukenspace/ratelimit"
)

func BenchmarkLimiterGroup(b *testing.B) {
	rate := ratelimit.ReqPerSec(100 / 60)
	burst := 100
	limiterGroups := map[string]ratelimit.LimiterGroup{
		"SyncMap + LoadOrStore":        &ratelimit.LiGrSyncMapLoadOrStore{R: rate, B: burst},
		"SyncMap + Load > LoadOrStore": &ratelimit.LiGrSyncMapLoadThenLoadOrStore{R: rate, B: burst},
		"SyncMap + Load > Store":       &ratelimit.LiGrSyncMapLoadThenStore{R: rate, B: burst},
		"Map + Mutex":                  ratelimit.NewLiGrMutex(rate, burst),
		"Map + RWMutex":                ratelimit.NewLiGrRWMutex(rate, burst),
	}

	for _, concurrentUsers := range []int{1028, 2046, 8192, 16384} {
		runtime.GOMAXPROCS(16)
		for name, limiter := range limiterGroups {
			limiterGroupRunner(b, name, concurrentUsers, limiter)
		}
	}
}

func limiterGroupRunner(b *testing.B, name string, numberOfKeys int, limiter ratelimit.LimiterGroup) bool {
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
			}(strconv.Itoa(key), batch)
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
