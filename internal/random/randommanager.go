package random

import (
	"math/rand"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func MakeRandomNumber(min, max int) int {
	return rnd.Intn(max-min) + min
}
