# go-ratelimit

## Benchmarking
CPU Info

```bash
> sysctl -a machdep.cpu
machdep.cpu.cores_per_package: 10
machdep.cpu.core_count: 10
machdep.cpu.logical_per_package: 10
machdep.cpu.thread_count: 10
machdep.cpu.brand_string: Apple M1 Pro
```

Results
| Benchmark                                                               | Iterations | Time per Operation (ns/op) | Memory per Operation (B/op) | Allocations per Operation (allocs/op) |
| ----------------------------------------------------------------------- | ---------- | -------------------------- | --------------------------- | ------------------------------------- |
| BenchmarkLiGr/name:SyncMap_+_Load_>_LoadOrStore;number_of_keys:8192-10  | 5273980    | 329.5 ns/op                | 89 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_Load_>_Store;number_of_keys:8192-10        | 6777685    | 413.4 ns/op                | 111 B/op                    | 1 allocs/op                           |
| BenchmarkLiGr/name:Map_+_Mutex;number_of_keys:8192-10                   | 1843262    | 704.6 ns/op                | 48 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:Map_+_RWMutex;number_of_keys:8192-10                 | 5528402    | 628.6 ns/op                | 148 B/op                    | 1 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_LoadOrStore;number_of_keys:8192-10         | 5929664    | 2905 ns/op                 | 346 B/op                    | 3 allocs/op                           |
| BenchmarkLiGr/name:Map_+_Mutex;number_of_keys:16384-10                  | 1808432    | 660.3 ns/op                | 49 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:Map_+_RWMutex;number_of_keys:16384-10                | 3640098    | 449.9 ns/op                | 48 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_LoadOrStore;number_of_keys:16384-10        | 3183156    | 377.0 ns/op                | 145 B/op                    | 3 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_Load_>_LoadOrStore;number_of_keys:16384-10 | 4056458    | 747.7 ns/op                | 48 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_Load_>_Store;number_of_keys:16384-10       | 3867928    | 512.0 ns/op                | 48 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_LoadOrStore;number_of_keys:32768-10        | 4796886    | 380.3 ns/op                | 145 B/op                    | 3 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_Load_>_LoadOrStore;number_of_keys:32768-10 | 4463239    | 353.3 ns/op                | 49 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:SyncMap_+_Load_>_Store;number_of_keys:32768-10       | 3094686    | 328.0 ns/op                | 50 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:Map_+_Mutex;number_of_keys:32768-10                  | 1600137    | 697.6 ns/op                | 52 B/op                     | 1 allocs/op                           |
| BenchmarkLiGr/name:Map_+_RWMutex;number_of_keys:32768-10                | 5107765    | 596.4 ns/op                | 49 B/op                     | 1 allocs/op                           |
