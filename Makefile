.PHONY: test bench-all bench-limiter bench-ratelimit

TEST_COUNT?=1
TEST_BENCHTIME?=10s

test:
	go test ./... -v -parallel=8

bench-all: bench-limiter bench-ratelimit
bench-ratelimit:
	for time in 10s 30s; do \
		TEST_BENCHTIME=$$time $(MAKE) bench-isolated; \
	done
	for time in 10s 30s; do \
		TEST_BENCHTIME=$$time $(MAKE) bench-distributed; \
	done

bench-limiter:
	go test ./limiter/... -v -bench=. -run=^$ -test.count=$(TEST_COUNT) -test.benchtime=$(TEST_BENCHTIME)  > out/bench/limiter/$(TEST_BENCHTIME).txt

bench-isolated:
	go test ./v1/ratelimit/... -v -bench=BenchmarkIsolated -run=^$ -test.count=$(TEST_COUNT) -test.benchtime=$(TEST_BENCHTIME) > out/bench/ratelimit/isolated/$(TEST_BENCHTIME).txt

bench-distributed:
	go test ./v1/ratelimit/... -v -bench=BenchmarkDistributed -run=^$ -test.count=$(TEST_COUNT) -test.benchtime=$(TEST_BENCHTIME) > out/bench/ratelimit/distributed/$(TEST_BENCHTIME).txt