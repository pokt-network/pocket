package raintree

import (
	"math/rand"
	"time"
)

func GenerateRandInt() int32 {
	rand.Seed(time.Now().Unix())
	return rand.Int31()
}
