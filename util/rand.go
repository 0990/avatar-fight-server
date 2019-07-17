package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Rand32() int32 {
	return rand.Int31()
}

func Randn(n int) int {
	return rand.Intn(n)
}

func Rand32n(n int32) int32 {
	return rand.Int31n(n)
}

func Rand64() int64 {
	return rand.Int63()
}

func Rand64n(n int64) int64 {
	return rand.Int63n(n)
}
