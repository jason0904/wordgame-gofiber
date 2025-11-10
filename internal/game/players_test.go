package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetNextPlayerTurn(t *testing.T) {
	g := SetupDefaultPlayers()

	g.setNextPlayerTurn("1001")
	assert.Equal(t, "1002", g.currentUserID, "Next turn should be Bob's ID")

	g.setNextPlayerTurn("1002")
	assert.Equal(t, "1003", g.currentUserID, "Next turn should be Charlie's ID")

	g.setNextPlayerTurn("1003")
	assert.Equal(t, "1001", g.currentUserID, "Next turn should wrap around to Alice's ID")
}

func TestAddUser(t *testing.T) {
	g := SetupTestGame()

	user := &User{ID: "u1", Name: "Alice"}
	g.addUser(user)

	if len(g.players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(g.players))
	}
}

func TestRemoveUser(t *testing.T) {
	g := SetupDefaultPlayers()

	user1 := g.players[0] // Alice
	user2 := g.players[1] // Bob

	g.removeUser(user1)

	assert.NotContains(t, g.players, user1, "Player list should not contain the removed user")
	assert.Contains(t, g.players, user2, "Player list should still contain other users")
}
