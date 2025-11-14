package game

import (
	"log"
	"strconv"
)

// ...existing code...
func (g *Game) generateUniqueID() string {
	const maxAttempts = MAXIDENTIFIER - MINIDENTIFIER + 1
	for i := 0; i < maxAttempts; i++ {
		id := g.makeRandomID()
		exists := g.isIDExist(id)
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

func (g *Game) makeRandomID() string {
	n := g.random.MakeRandomNumber(MINIDENTIFIER, MAXIDENTIFIER+1)
	return strconv.Itoa(n)
}

func (g *Game) isIDExist(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isIDInPlayers(id) {
		return true
	}
	if g.isIDInSpectators(id) {
		return true
	}
	return false
}

func (g *Game) isIDInPlayers(id string) bool {
	for _, p := range g.players {
		if g.isIDEqual(p.ID, id) {
			return true
		}
	}
	return false
}

func (g *Game) isIDInSpectators(id string) bool {
	for _, s := range g.spectators {
		if g.isIDEqual(s.ID, id) {
			return true
		}
	}
	return false
}

func (g *Game) isIDEqual(id1, id2 string) bool {
	return id1 == id2
}
