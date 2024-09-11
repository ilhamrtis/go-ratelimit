.PHONY: test bench

test:
	go test ./... -v -parallel=8

bench-all: bench-limiter bench-limiter-group

bench-limiter:
	go test limiter_bench_test.go -v -bench=. -test.count=3 > out/bench/limiter.txt 
bench-limiter-group:
	go test limiter_group_bench_test.go -v -bench=. -test.count=3 -test.benchtime=1s > out/bench/limiter_group_1s.txt
	go test limiter_group_bench_test.go -v -bench=. -test.count=3 -test.benchtime=2s > out/bench/limiter_group_2s.txt
	go test limiter_group_bench_test.go -v -bench=. -test.count=3 -test.benchtime=5s > out/bench/limiter_group_5s.txt