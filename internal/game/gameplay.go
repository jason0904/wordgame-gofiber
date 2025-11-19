package game

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

func (g *Game) handlePlay(user *User, word string) {
	g.mu.Lock()

	if g.handleGameAlreadyStarted() {
		return
	}

	if g.handleUserIsNotCurrentTurn(user.ID) {
		return
	}

	word = strings.TrimSpace(word)
	if g.handleWordIsBlank(word) {
		return
	}
	if g.handleWordIsNotEnoughLength(word) {
		return
	}
	if g.handleWordIsAlreadyUsed(user, word) {
		return
	}
	if g.handleWordChainRuleDismatch(user, word) {
		return
	}
	if g.handleWordIsNotInDB(user, word) {
		return
	}

	g.handleNextTurn(user, word)
}

func (g *Game) endGame(message string) {
	g.mu.Lock()
	g.gameover = true
	g.message = message
	g.mu.Unlock()
	log.Printf(RESETLOGMSG, g.RoomId)
	g.broadcastGameState()

	//5초 후에 게임 리셋
	go func() {
		time.Sleep(5 * time.Second)
		g.mu.Lock()
		g.reset()
		g.mu.Unlock()
		g.broadcastGameState()
	}()
}

func (g *Game) reset() {
	if g.gameover {
		g.players = append(g.players, g.spectators...)
		g.spectators = make([]*User, 0)
	}

	g.startword = ""
	g.lastWord = ""
	g.usedWords = make(map[string]bool)
	g.gameover = false
	g.started = false
	g.currentUserID = ""
	g.message = AVAILABLEMSG
	log.Printf(RESETLOGMSG, g.RoomId)
}

func (g *Game) startGame(user *User) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.started {
		g.message = GAMEALREADYSTARTEDMSG
		return
	} else if g.hostUserId != user.ID {
		g.message = NOHOSTPRIVILEGESMSG
		return
	} else if len(g.players) < MinPlayersToStart {
		g.message = fmt.Sprintf(MINPLAYERTOSTARTMSG, MinPlayersToStart)
		return
	}

	g.startNewRound()
}

func (g *Game) startNewRound() {
	if len(g.players) == 0 {
		g.reset()
		return
	}

	randomPlayerIndex := g.selectRandomPlayerIndex()
	g.startword = g.makeStartWord()
	g.lastWord = g.startword
	g.usedWords = make(map[string]bool)
	g.usedWords[g.startword] = true
	g.started = true
	g.gameover = false
	g.currentUserID = g.players[randomPlayerIndex].ID
	g.message = fmt.Sprintf(STARTMSG, g.makeNameToDisplay(g.currentUserID, g.players[randomPlayerIndex].Name))
	log.Printf(STARTLOGMSG, g.RoomId)
}

func (g *Game) eliminatePlayer(user *User, reason string) (winner bool, winnerMsg string) {
	eliminated := false

	for i, p := range g.players {
		g.handleUserElimination(p, user, i, reason)
	}

	winner, msg := g.handleWinnnerCheck()
	if winner {
		return true, msg
	}

	if eliminated {
		g.startNewRound()
	}

	return false, ""
}

func (g *Game) makeStartWord() string {
	randomWordLength := g.random.MakeRandomNumber(MinStartWordLength, MaxStartWordLength) // 2자에서 6자 사이
	word, err := g.store.GetRandomWordByLength(randomWordLength)
	if err != nil {
		log.Println(STARTINGWORDERRORLOGMSG, err)
		return NORMALSTARTWORD
	}
	return word
}

func (g *Game) wordDBCheck(word string) bool {
	return g.store.IsWordInDB(word)
}

func (g *Game) handleGameAlreadyStarted() bool {
	if !g.started {
		g.message = NOTTOHANDLEPLAYMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func (g *Game) handleUserIsNotCurrentTurn(id string) bool {
	if g.currentUserID != id {
		g.message = NOTCURRENTPLAYERSMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func (g *Game) handleWordIsBlank(word string) bool {
	if word == "" {
		g.message = TYPEWORDMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func (g *Game) handleWordIsNotEnoughLength(word string) bool {
	if utf8.RuneCountInString(word) < 2 {
		g.message = MINWORDLENGTHMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func (g *Game) handleEndGameOrContinue(winner bool, msg string) {
	if winner {
		g.endGame(msg)
	} else {
		g.broadcastGameState()
	}
}

func (g *Game) handleWordIsAlreadyUsed(user *User, word string) bool {
	if g.usedWords[word] {
		winner, msg := g.eliminatePlayer(user, WORDALREADYUSEDMSG)
		g.mu.Unlock()
		g.handleEndGameOrContinue(winner, msg)
		return true
	}
	return false
}

func (g *Game) handleWordChainRuleDismatch(user *User, word string) bool {
	lastRune, _ := utf8.DecodeLastRuneInString(g.lastWord)
	firstRune, _ := utf8.DecodeRuneInString(word)
	if lastRune != firstRune {
		winner, msg := g.eliminatePlayer(user, WORDMISMATCHMSG)
		g.mu.Unlock()
		g.handleEndGameOrContinue(winner, msg)
		return true
	}
	return false
}

func (g *Game) handleWordIsNotInDB(user *User, word string) bool {
	if !g.wordDBCheck(word) {
		winner, msg := g.eliminatePlayer(user, WORDNOTINDICTMSG)
		g.mu.Unlock()
		g.handleEndGameOrContinue(winner, msg)
		return true
	}
	return false
}

func (g *Game) handleNextTurn(user *User, word string) {
	g.lastWord = word
	g.usedWords[word] = true
	g.setNextPlayerTurn(user.ID)
	g.mu.Unlock()
	g.broadcastGameState()
}

func (g *Game) handleUserElimination(user *User, target *User, index int, reason string) {
	if target.ID == user.ID {
		g.players = append(g.players[:index], g.players[index+1:]...)
		g.spectators = append(g.spectators, user)
		g.message = g.makeNameToDisplay(user.ID, user.Name) + ELIMINATEDMSG + reason
	}
}

func (g *Game) handleWinnnerCheck() (bool, string) {
	// 승리 조건: 활성 플레이어가 한 명이면 우승 처리 (잠금은 호출자가 관리)
	if len(g.players) == 1 {
		winner := g.players[0]
		msg := g.makeNameToDisplay(winner.ID, winner.Name) + WINNERMSG
		g.gameover = true
		g.message = msg
		return true, msg
	}

	return false, ""
}
