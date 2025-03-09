package math

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomFloat() float64 {
	return rand.Float64()
}

func RandomFloatRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func RandomBool() bool {
	return rand.Float64() < 0.5
}

func SetRandomSeed(seed int64) {
	rand.Seed(seed)
} 