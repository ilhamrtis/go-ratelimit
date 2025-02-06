# go-ratelimit

`go-ratelimit` is a Go library for rate limiting using various synchronization mechanisms. It provides different implementations of rate limiters using `sync.Mutex`, `sync.RWMutex`, and `sync.Map`.

## Project Status

This project is in BETA. Contributions and suggestions are welcome.

## Features

- A low impact distributed rate limiter [./v1/ratelimit/redis_delayed_sync.go]

## Installation

To install the library, use `go get`:

```bash
go get -u github.com/yesyoukenspace/ratelimit
```

## Usage
All `limiter.Limiter` implementations follow the `Ratelimit` interface in [domain.go](./domain.go) 

```go
package main

import (
	"fmt"
	"github.com/yesyoukenspace/go-ratelimit/v1"
)

func main() {
	ratelimiter := ratelimit.NewDefault(10, 5)
	key := "user123"
	if ok, _ := ratelimiter.Allow(key); ok {
		fmt.Println("Request allowed")
	} else {
		fmt.Println("Request denied")
	}
}
```

### Available Implementations
Here are the available implementations:

- **Mutex**: Uses `sync.Mutex`.
- **RWMutex**: Uses `sync.RWMutex`.
- **SyncMapLoadThenLoadOrStore**: Uses `sync.Map` with `Load` then `LoadOrStore`.
- **SyncMapLoadOrStore**: Uses `sync.Map` with `LoadOrStore`.
- **SyncMapLoadThenStore**: Uses `sync.Map` with `Load` then `Store`.

Refer to the respective files for implementation details:
- [mutex.go](mutex.go)
- [sync_map.go](sync_map.go)

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

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Authors and Acknowledgments


