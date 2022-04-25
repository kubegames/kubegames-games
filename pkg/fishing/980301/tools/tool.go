package tools

import (
	"math/rand"
	"time"
)

var (
	lastTime = int64(0)
	diff = int64(0)
)

func RandInt (min, max int, i int64) int {
	if max - min <= 0 {return min}
	now := time.Now().UnixNano()
	if now == lastTime {
		diff++
		i += diff
	} else {
		diff = 0
	}
	lastTime = now
	rand.Seed(now + i * 1000)
	return rand.Intn(max - min) + min
}
