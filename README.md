# go-ratelimit

`go-ratelimit` is a Go library for rate limiting using various synchronization mechanisms. It provides different implementations of rate limiters using `sync.Mutex`, `sync.RWMutex`, and `sync.Map`.

## Features
- All `limiter.Limiter` implementations follow the `Ratelimit` interface in [domain.go](./domain.go) which closely follows `golang.org/x/time/rate` interface as much as possible.

## Installation

To install the library, use `go get`:

```bash
go get -u github.com/yesyoukenspace/go-ratelimit
```

## Usage

For usage examples checkout [distributed_bench_test.go](v1/ratelimit/distributed_bench_test.go) 

## Implementations - Distributed Rate Limiting 
Look at [distributed_bench_test.go](v1/ratelimit/distributed_bench_test.go) for examples.

### **RedisDelayedSync**
- A low impact rate limiter backed by redis made for distributed systems that requires low latency but allows some inaccuracy in enforcing rate limit - [code](./v1/ratelimit/redis_delayed_sync.go)
- Uses `SyncMapLoadThenStore` to manage local states and synchronizes with redis asynchronously thus not blocking `AllowN` functions

### **GoRedisRate**
This is just a wrapper around `github.com/go-redis/redis_rate`, used primarily for testing and benchmarking.

## Implementations - Isolated Rate Limiting 
- **Mutex**: Uses `sync.Mutex`.
- **RWMutex**: Uses `sync.RWMutex`.
- **SyncMapLoadThenLoadOrStore**: Uses `sync.Map` with `Load` then `LoadOrStore`.
- **SyncMapLoadOrStore**: Uses `sync.Map` with `LoadOrStore`.
- **SyncMapLoadThenStore**: Uses `sync.Map` with `Load` then `Store`.

## Benchmarking
Benchmarking results at [./out/bench](./out/bench/)


### Benchmarking results
Info of machine 
```bash
> sysctl -a machdep.cpu
machdep.cpu.cores_per_package: 10
machdep.cpu.core_count: 10
machdep.cpu.logical_per_package: 10
machdep.cpu.thread_count: 10
machdep.cpu.brand_string: Apple M1 Pro
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change. Make sure to update tests as appropriate.

## License

See the [LICENSE](LICENSE) file for details.

## References
- [go-redis/redis_rate](https://github.com/go-redis/redis_rate)
- [uber-go/ratelimit](https://github.com/uber-go/ratelimit)

## Authors and Acknowledgments
- @YesYouKenSpace


