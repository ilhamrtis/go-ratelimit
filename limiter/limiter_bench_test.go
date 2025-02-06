package limiter

import (
	"fmt"
	"math"
	"sync/atomic"

	"testing"
)

func BenchmarkLimiter(b *testing.B) {
	r := math.Pow10(6)
	burst := int(math.Pow10(1))
	limiters := []struct {
		name    string
		limiter Limiter
	}{
		{name: "Bucket", limiter: NewBucket(r, burst)},
		{name: "DefaultLimiter", limiter: NewDefaultLimiter(r, burst)},
		{name: "ResetBasedLimiter", limiter: NewResetbasedLimiter(r, burst)},
	}

	for _, numberOfGoroutines := range []int{1, 8, 32} {
		for _, limiter := range limiters {
			limiterRunner(b, limiter.name, numberOfGoroutines, limiter.limiter)
		}
	}
}

func limiterRunner(b *testing.B, name string, numberOfGoroutines int, limiter Limiter) bool {
	maxLapsed, maxAllowed, maxDisallowed := 0.0, 0.0, 0.0
	runResult := b.Run(fmt.Sprintf("%s/goroutines=%d", name, numberOfGoroutines), func(b *testing.B) {
		b.ReportAllocs()
		b.SetParallelism(numberOfGoroutines)
		allowed := atomic.Int64{}
		disallowed := atomic.Int64{}
		b.RunParallel(func(pb *testing.PB) {
			a, d := int64(0), int64(0)
			for pb.Next() {
				if ok, _ := limiter.Allow(); ok {
					a++
				} else {
					d++
				}
			}
			allowed.Add(a)
			disallowed.Add(d)
		})

		lapsed := b.Elapsed().Seconds()
		if lapsed > maxLapsed {
			maxLapsed = lapsed
			maxAllowed = float64(allowed.Load())
			maxDisallowed = float64(disallowed.Load())
		}
	})
	fmt.Printf("%s/custom_metrics\tallowed=%f\tdisallowed=%f\tlapsed=%f\tallowed(rps)=%f\ttotal(rps)=%f\n", b.Name()+"/"+name, maxAllowed, maxDisallowed, maxLapsed, maxAllowed/maxLapsed, (maxAllowed+maxDisallowed)/maxLapsed)
	return runResult
}
