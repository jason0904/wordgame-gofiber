package game

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"unicode/utf8"
)

func TestWordDBCheck(t *testing.T) {
	g := SetupTestGame()

	validWord := "사과"
	invalidWord := "asdfgh"

	assert.True(t, g.wordDBCheck(validWord), "Valid word should be found in the database")
	assert.False(t, g.wordDBCheck(invalidWord), "Invalid word should not be found in the database")
}

func TestMakeStartWord(t *testing.T) {
	g := SetupTestGame()

	startWord := g.makeStartWord()

	assert.GreaterOrEqual(t, utf8.RuneCountInString(startWord), MinStartWordLength, "Start word should be at least MinStartWordLength characters long")
	assert.Less(t, utf8.RuneCountInString(startWord), MaxStartWordLength, "Start word should be less than  MaxStartWordLength characters long")
	assert.True(t, g.wordDBCheck(startWord), "Start word should be valid according to the word database")
}

func TestNewRound(t *testing.T) {
	g := SetupDefaultPlayers()

	g.startNewRound()

	assert.Equal(t, 1, len(g.usedWords), "Used words should be reset for new round") //시작 단어는 들어가 있어야하니.
	assert.True(t, g.started, "Game should be marked as started")
	assert.False(t, g.gameover, "Game should not be marked as over")
	assert.NotEmpty(t, g.currentUserID, "Current user ID should be set to a player")
	assert.Equal(t, g.startword, g.lastWord, "Last word should be set to start word")
	assert.Contains(t, g.usedWords, g.startword, "Start word should be in used words")
}

func TestStartGame(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]

	g.startGame(host)

	assert.True(t, g.started, "Game should be marked as started")
}

func TestStartGameNotStartedAgain(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]

	g.startGame(host)
	g.startGame(host)

	assert.Equal(t, g.message, GAMEALREADYSTARTEDMSG, "Message should indicate game has already started")
}

func TestStartGameNotHost(t *testing.T) {
	g := SetupDefaultPlayers()

	nonHost := g.players[1]

	g.startGame(nonHost)

	assert.Equal(t, g.message, NOHOSTPRIVILEGESMSG, "Message should indicate lack of host privileges")
}

func TestStartGameMinPlayerNotEntered(t *testing.T) {
	g := SetupTestGame()

	host := &User{ID: "host1", Name: "Host"}

	g.addUser(host)
	g.startGame(host)

	expectedMsg := fmt.Sprintf(MINPLAYERTOSTARTMSG, MinPlayersToStart)
	assert.Equal(t, g.message, expectedMsg, "Message should indicate not enough players to start")
}

func TestReset(t *testing.T) {

	g := SetupDefaultPlayers()

	host := g.players[0]

	g.startGame(host)
	g.reset()

	assert.False(t, g.started, "Game should be marked as not started")
	assert.False(t, g.gameover, "Game should be marked as not over")
	assert.Empty(t, g.currentUserID, "Current user ID should be cleared")
	assert.Empty(t, g.startword, "Start word should be cleared")
	assert.Empty(t, g.lastWord, "Last word should be cleared")
	assert.Empty(t, g.usedWords, "Used words should be cleared")
}

func TestEndGame(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]

	g.startGame(host)
	beforeGoroutines := runtime.NumGoroutine()
	g.endGame("test game over")

	assert.True(t, g.gameover, "Game should be marked as over")
	assert.Equal(t, g.message, "test game over", "Message should be set to game over message")
	afterGoroutines := runtime.NumGoroutine()
	assert.Equal(t, 1, afterGoroutines-beforeGoroutines, "There should be one additional goroutine for resetting the game")
}

func TestEliminatePlayer(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]

	g.startGame(host)
	playerToEliminate := g.players[1]
	g.eliminatePlayer(playerToEliminate, "test elimination")

	assert.Equal(t, 2, len(g.players), "There should be 2 players left after elimination")
	assert.Equal(t, 1, len(g.spectators), "There should be 1 spectator after elimination")
	assert.Equal(t, g.spectators[0].ID, playerToEliminate.ID, "Eliminated player should be in spectators list")
}

func TestHandlePlayIfGameNotStarted(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]
	g.currentUserID = host.ID
	g.handlePlay(host, "사과")

	assert.Equal(t, g.message, NOTTOHANDLEPLAYMSG, "Message should indicate game has not started")
}

func TestHandlePlayIfNotCurrentUserEntered(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]
	g.startGame(host)
	g.currentUserID = host.ID
	nonCurrentPlayer := g.players[1]
	g.handlePlay(nonCurrentPlayer, "사과")

	assert.Equal(t, g.message, NOTTOHANDLEPLAYMSG, "Message should indicate not current user's turn")
}

func TestHandlePlayIfWordNotEntered(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]
	g.startGame(host)
	g.currentUserID = host.ID
	g.handlePlay(host, "")

	assert.Equal(t, g.message, TYPEWORDMSG, "Message should prompt to enter a word")
}

func TestHandlePlayIfWordTooShort(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]
	g.startGame(host)
	g.currentUserID = host.ID
	g.handlePlay(host, "사")

	assert.Equal(t, g.message, MINWORDLENGTHMSG, "Message should indicate word is too short")
}

func TestHandlePlayIfWordIsNotMatchRule(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]
	g.startGame(host)
	g.currentUserID = host.ID
	g.lastWord = "사과"
	g.handlePlay(host, "바나나")

	assert.Equal(t, len(g.spectators), 1, "One player should be eliminated")
}

func TestHandlePlayIfWordIsNotInDict(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]
	g.startGame(host)
	g.currentUserID = host.ID
	g.lastWord = "사과"
	g.handlePlay(host, "과과과과")

	assert.Equal(t, len(g.spectators), 1, "One player should be eliminated")
}

func TestHandlePlayIfWordAlreadyUsed(t *testing.T) {
	g := SetupDefaultPlayers()

	host := g.players[0]
	g.startGame(host)
	g.currentUserID = host.ID
	g.lastWord = "사과"
	g.usedWords["과일"] = true
	g.handlePlay(host, "과일")

	assert.Equal(t, len(g.spectators), 1, "One player should be eliminated")
}
