package ratelimit_test

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
