.PHONY: test bench

TEST_COUNT?=3
TEST_BENCHTIME?=1s

test:
	go test ./... -v -parallel=8

bench-all: bench-limiter
	TEST_BENCHTIME="1s" $(MAKE) bench-ratelimit
	TEST_COUNT=1 TEST_BENCHTIME="10s" $(MAKE) bench-ratelimit

bench-limiter:
	go test ./v1/ratelimit/ratelimit_bench_test.go -v -bench=. -test.count=$(TEST_COUNT) > out/bench/limiter/limiter.txt 

bench-ratelimit:
	go test ./v1/ratelimit/ratelimit_bench_test.go -v -bench=. -test.count=$(TEST_COUNT) -test.benchtime=$(TEST_BENCHTIME) > out/bench/ratelimit/$(TEST_BENCHTIME).txt