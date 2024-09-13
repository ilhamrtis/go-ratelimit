package ratelimit_test

import (
	"crypto/rand"
	"encoding/base64"
)

func isCloseEnough(a, b, tolerance float64) bool {
	return a >= b-tolerance && a <= b+tolerance
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func randString(bytes int) string {
	token := make([]byte, bytes)
	rand.Read(token)
	return base64.StdEncoding.EncodeToString(token)
}
