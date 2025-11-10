package game

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestGenerateUniqueID(t *testing.T) {

	g := SetupTestGame()

	id1 := g.generateUniqueID()
	id2 := g.generateUniqueID()

	id1ToInt, err1 := strconv.Atoi(id1)
	id2ToInt, err2 := strconv.Atoi(id2)

	assert.NoError(t, err1, "ID1 should be convertible to integer")
	assert.NoError(t, err2, "ID2 should be convertible to integer")
	assert.Greater(t, id1ToInt, 999, "Generated ID1 should be at least 1000")
	assert.Greater(t, id2ToInt, 999, "Generated ID2 should be at least 1000")
	assert.Less(t, id1ToInt, 10000, "Generated ID1 should be at most 9999")
	assert.Less(t, id2ToInt, 10000, "Generated ID2 should be at most 9999")

	assert.NotEqual(t, id1, id2, "Generated IDs should be unique")

}

func TestSelectRandomPlayerIndex(t *testing.T) {
	testGame := SetupDefaultPlayers()

	index := testGame.selectRandomPlayerIndex()
	assert.GreaterOrEqual(t, index, 0, "Selected index should be at least 0")
	assert.Less(t, index, len(testGame.players), "Selected index should be less than number of players")
}

func TestMakeNameToDisplay(t *testing.T) {
	g := SetupTestGame()
	userID := "1234"
	userName := "PlayerOne"
	expectedDisplayName := "PlayerOne#1234"

	displayName := g.makeNameToDisplay(userID, userName)
	assert.Equal(t, expectedDisplayName, displayName, "Display name should be correctly formatted")
}
