package random

import (
	"math/rand"
	"time"
)

type Manager struct {
	rnd *rand.Rand
}

func NewManager() *Manager {
	return &Manager{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *Manager) MakeRandomNumber(min, max int) int {
	return r.rnd.Intn(max-min) + min
}
