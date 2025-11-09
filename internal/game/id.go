package game

import (
	"log"
	"strconv"
)

// ...existing code...
func (g *Game) generateUniqueID() string {
	const maxAttempts = MAXIDENTIFIER - MINIDENTIFIER + 1
	for i := 0; i < maxAttempts; i++ {
		n := g.random.MakeRandomNumber(MINIDENTIFIER, MAXIDENTIFIER+1)
		id := strconv.Itoa(n)

		g.mu.Lock()
		exists := false
		for _, p := range g.players {
			if p.ID == id {
				exists = true
				break
			}
		}
		if !exists {
			for _, s := range g.spectators {
				if s.ID == id {
					exists = true
					break
				}
			}
		}
		g.mu.Unlock()

		if !exists {
			return id
		}
	}
	log.Println(IDMAXATTEMPTSLOGMSG)
	return IDFALLBACK
}

func (g *Game) selectRandomPlayerIndex() int {
	return g.random.MakeRandomNumber(0, len(g.players))
}

func (g *Game) makeNameToDisplay(userID, userName string) string {
	return userName + IDSUFFIX + userID
}