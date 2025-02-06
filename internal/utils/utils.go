package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math"
	"math/big"
)

func IsCloseEnough(a, b, tolerance float64) bool {
	if tolerance < 0 || tolerance > 0.5 {
		panic("tolerance must be between 0 and 0.5")
	}
	return a >= math.Floor(b*(1-tolerance)) && a <= math.Ceil(b*(1+tolerance))
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

func RandInt(min, max int64) int64 {
	if val, err := rand.Int(rand.Reader, big.NewInt(max-min)); err == nil {
		return min + val.Int64()
	} else {
		panic(err)
	}
}
func ExpectedAllowedRequests(burst float64, reqPerSec float64, seconds float64) int {
	return int(reqPerSec*seconds) + int(burst)
}
