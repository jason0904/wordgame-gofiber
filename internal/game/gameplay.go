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

	if handleGameAlreadyStarted(g) {
		return
	}

	if handleUserIsNotHost(g, user.ID) {
		return
	}

	word = strings.TrimSpace(word)
	if handleWordIsBlank(g, word) {
		return
	}
	if handleWordIsNotEnoughLength(g, word) {
		return
	}
	if handleWordIsAlreadyUsed(g, user, word) {
		return
	}
	if handleWordChainRuleDismatch(g, user, word) {
		return
	}
	if handleWordIsNotInDB(g, user, word) {
		return
	}

	handleNextTurn(g, user, word)
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
		handleUserElimination(g, p, user, i, reason)
	}

	winner, msg := handleWinnnerCheck(g)
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

// 비공개 메서드

func handleGameAlreadyStarted(g *Game) bool {
	if !g.started {
		g.message = NOTTOHANDLEPLAYMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func handleUserIsNotHost(g *Game, id string) bool {
	if g.hostUserId != id {
		g.message = NOTCURRENTPLAYERSMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func handleWordIsBlank(g *Game, word string) bool {
	if word == "" {
		g.message = TYPEWORDMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func handleWordIsNotEnoughLength(g *Game, word string) bool {
	if utf8.RuneCountInString(word) < 2 {
		g.message = MINWORDLENGTHMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return true
	}
	return false
}

func handleEndGameOrContinue(g *Game, winner bool, msg string) {
	if winner {
		g.endGame(msg)
	} else {
		g.broadcastGameState()
	}
}

func handleWordIsAlreadyUsed(g *Game, user *User, word string) bool {
	if g.usedWords[word] {
		winner, msg := g.eliminatePlayer(user, WORDALREADYUSEDMSG)
		g.mu.Unlock()
		handleEndGameOrContinue(g, winner, msg)
		return true
	}
	return false
}

func handleWordChainRuleDismatch(g *Game, user *User, word string) bool {
	lastRune, _ := utf8.DecodeLastRuneInString(g.lastWord)
	firstRune, _ := utf8.DecodeRuneInString(word)
	if lastRune != firstRune {
		winner, msg := g.eliminatePlayer(user, WORDMISMATCHMSG)
		g.mu.Unlock()
		handleEndGameOrContinue(g, winner, msg)
		return true
	}
	return false
}

func handleWordIsNotInDB(g *Game, user *User, word string) bool {
	if !g.wordDBCheck(word) {
		winner, msg := g.eliminatePlayer(user, WORDNOTINDICTMSG)
		g.mu.Unlock()
		handleEndGameOrContinue(g, winner, msg)
		return true
	}
	return false
}

func handleNextTurn(g *Game, user *User, word string) {
	g.lastWord = word
	g.usedWords[word] = true
	g.setNextPlayerTurn(user.ID)
	g.mu.Unlock()
	g.broadcastGameState()
}

func handleUserElimination(g *Game, user *User, target *User, index int, reason string) {
	if target.ID == user.ID {
		g.players = append(g.players[:index], g.players[index+1:]...)
		g.spectators = append(g.spectators, user)
		g.message = g.makeNameToDisplay(user.ID, user.Name) + ELIMINATEDMSG + reason
	}
}

func handleWinnnerCheck(g *Game) (bool, string) {
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
