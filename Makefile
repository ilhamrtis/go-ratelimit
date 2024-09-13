.PHONY: test bench

TEST_COUNT?=3
TEST_BENCHTIME?=1s

test:
	go test ./... -v -parallel=8

bench-all: bench-limiter
	TEST_BENCHTIME="1s" $(MAKE) bench-limiter-group
	TEST_COUNT=1 TEST_BENCHTIME="10s" $(MAKE) bench-limiter-group

bench-limiter:
	go test limiter_bench_test.go -v -bench=. -test.count=$(TEST_COUNT) > out/bench/limiter.txt 

bench-limiter-group:
	go test limiter_group_bench_test.go utils_test.go -v -bench=. -test.count=$(TEST_COUNT) -test.benchtime=$(TEST_BENCHTIME) > out/bench/limiter_group_$(TEST_BENCHTIME).txt