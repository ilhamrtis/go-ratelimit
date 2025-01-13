package limiter

import (
	"fmt"
	"runtime"
	"sync"

	"testing"
)

func BenchmarkLimiter(b *testing.B) {
	r := (100.0 / 60)
	burst := 100
	limiters := map[string]Limiter{
		"Bucket":            NewBucket(r, burst),
		"DefaultLimiter":    NewDefaultLimiter(r, burst),
		"ResetBasedLimiter": NewResetbasedLimiter(r, burst),
	}

	for _, numberOfProcs := range []int{2, 4, 8} {
		runtime.GOMAXPROCS(numberOfProcs)
		for name, limiter := range limiters {
			limiterRunner(b, name, numberOfProcs, limiter)
		}
	}
}

func limiterRunner(b *testing.B, name string, numberOfProcs int, limiter Limiter) bool {
	return b.Run(fmt.Sprintf("name:%s;number of procs:%d", name, numberOfProcs), func(b *testing.B) {
		b.ReportAllocs()

		n := b.N
		batchSize := n / numberOfProcs
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
					if ok, _ := limiter.Allow(); ok {
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
