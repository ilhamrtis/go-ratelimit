package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func IsCloseEnough(a, b, tolerance float64) bool {
	return a >= b-tolerance && a <= b+tolerance
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func RandString(bytes int) string {
	token := make([]byte, bytes)
	rand.Read(token)
	return base64.StdEncoding.EncodeToString(token)
}
