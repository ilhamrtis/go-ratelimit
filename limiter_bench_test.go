package ratelimit_test

import (
	"fmt"
	"runtime"
	"sync"

	"testing"

	"github.com/yesyoukenspace/ratelimit"
	"golang.org/x/time/rate"
)

func BenchmarkLimiter(b *testing.B) {
	R := ratelimit.ReqPerSec(100 / 60)
	burst := 100
	limiters := map[string]ratelimit.Limiter{
		"Bucket":       ratelimit.NewBucket(R, burst),
		"rate.Limiter": rate.NewLimiter(rate.Limit(R), burst),
	}

	for _, numberOfGoRoutine := range []int{2, 4, 8} {
		runtime.GOMAXPROCS(numberOfGoRoutine)
		for name, limiter := range limiters {
			limiterRunner(b, name, numberOfGoRoutine, limiter)
		}
	}
}

func limiterRunner(b *testing.B, name string, numberOfGoRoutine int, limiter ratelimit.Limiter) bool {
	return b.Run(fmt.Sprintf("name:%s;number of goroutines:%d", name, numberOfGoRoutine), func(b *testing.B) {
		b.ReportAllocs()

		n := b.N
		batchSize := n / numberOfGoRoutine
		var wg sync.WaitGroup
		if batchSize == 0 {
			batchSize = n
		}
		key := 0
		mu := sync.RWMutex{}
		mu.Lock()
		allowed := 0
		disallowed := 0

		for n > 0 {
			wg.Add(1)
			batch := min(n, batchSize)
			n -= batch
			go func(quota int) {
				mu.RLock()
				defer mu.RUnlock()
				for i := 0; i < quota; i++ {
					if limiter.Allow() {
						allowed++
					} else {
						disallowed++
					}
				}
				wg.Done()
			}(batch)
			key++
		}
		b.StartTimer()
		mu.Unlock()
		wg.Wait()
		b.StopTimer()
		runtime.KeepAlive(allowed)
		runtime.KeepAlive(disallowed)
	})
}
