package ratelimit

import (
	"context"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/yesyoukenspace/go-ratelimit/internal/utils"
)

func TestSyncLimits(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	values := []interface{}{10, 10, 1, 10}

	key := utils.RandString(4)
	v, err := syncLimits.Run(ctx, rdb, []string{fmt.Sprintf("rate:%s", key)}, values...).Result()
	if err != nil {
		t.Fatal(err)
	}
	values = v.([]interface{})
	expected := []interface{}{int64(10), int64(10), int64(1), int64(10)}
	if len(values) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), len(values))
	}
	for i := range values {
		if values[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, values)
		}
	}
}
