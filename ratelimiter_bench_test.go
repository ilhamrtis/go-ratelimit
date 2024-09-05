package ratelimit_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"yesyoukenspace/ratelimit"
)

// inspired by https://github.com/uber-go/ratelimit/blob/main/ratelimit_bench_test.go
func BenchmarkRateLimiter(b *testing.B) {
	for _, procs := range []int{4, 8, 16} {
		runtime.GOMAXPROCS(procs)
		for name, limiter := range map[string]ratelimit.Ratelimiter{
			"Default Limiter":       ratelimit.NewDefaultLimiter(100/60, 100),
			"LoadStore Limiter":     &ratelimit.SyncMapLimiterWithLoadStore{M: sync.Map{}, R: 100 / 60, B: 100},
			"LoadLoadStore Limiter": &ratelimit.SyncMapLimiterWithLoadLoadStore{M: sync.Map{}, R: 100 / 60, B: 100},
		} {
			for nk := 1; nk < procs; nk++ {
				runner(b, name, nk, limiter)
			}
		}
	}
}

func runner(b *testing.B, name string, numberOfKeys int, limiter ratelimit.Ratelimiter) bool {
	return b.Run(fmt.Sprintf("name:%s;number of keys:%d", name, numberOfKeys), func(b *testing.B) {
		b.ReportAllocs()
		var wgg sync.WaitGroup
		n := b.N
		batchSize := n / numberOfKeys
		if batchSize == 0 {
			batchSize = n
		}
		key := 0
		for n > 0 {
			batch := min(n, batchSize)
			n -= batch
			wgg.Add(1)
			go func(k string, quota int) {
				var wg sync.WaitGroup
				wg.Add(quota)

				for i := 0; i < quota; i++ {
					go func() {
						limiter.Allow(k)
						wg.Done()
					}()
				}
				wg.Wait()
				wgg.Done()
			}(fmt.Sprintf("key%d", key), batch)
			key++
		}
		b.StartTimer()
		wgg.Wait()
		b.StopTimer()
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
