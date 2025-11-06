package main

import (

	"time"
	"math/rand"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func makeRandomNumber(min, max int) int {
	return rnd.Intn(max-min) + min
}