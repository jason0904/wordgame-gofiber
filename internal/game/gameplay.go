package game

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

// ...existing code...
func (g *Game) handlePlay(user *User, word string) {
	g.mu.Lock()

	if g.gameover || g.currentUserID != user.ID {
		g.mu.Unlock()
		return
	}

	word = strings.TrimSpace(word)
	if word == "" {
		g.message = TYPEWORDMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return
	}

	if utf8.RuneCountInString(word) < 2 {
		g.message = MINWORDLENGTHMSG
		g.mu.Unlock()
		g.broadcastGameState()
		return
	}

	if g.usedWords[word] {
		winner, msg := g.eliminatePlayerUnlocked(user, WORDALREADYUSEDMSG)
		g.mu.Unlock()
		if winner {
			g.endGame(msg)
		} else {
			g.broadcastGameState()
		}
		return
	}

	if g.lastWord != "" {
		lastRune, _ := utf8.DecodeLastRuneInString(g.lastWord)
		firstRune, _ := utf8.DecodeRuneInString(word)
		if lastRune != firstRune {
			winner, msg := g.eliminatePlayerUnlocked(user, WORDMISMATCHMSG)
			g.mu.Unlock()
			if winner {
				g.endGame(msg)
			} else {
				g.broadcastGameState()
			}
			return
		} else if !g.wordDBCheck(word) {
			winner, msg := g.eliminatePlayerUnlocked(user, WORDNOTINDICTMSG)
			g.mu.Unlock()
			if winner {
				g.endGame(msg)
			} else {
				g.broadcastGameState()
			}
			return
		}
	}

	g.lastWord = word
	g.usedWords[word] = true
	g.setNextPlayerTurn(user.ID)
	g.mu.Unlock()
	g.broadcastGameState()
}

func (g *Game) endGame(message string) {
	g.mu.Lock()
	g.gameover = true
	g.message = message
	g.mu.Unlock()
	log.Printf(RESETLOGMSG, g.RoomId)
	g.broadcastGameState()

	// 5초 후 게임을 리셋하여 로비로 돌아감
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
		g.message = ALREADYSTARTEDMSG
		return
	}

	if g.hostUserId != user.ID {
		g.message = NOHOSTPRIVILEGESMSG
		return
	}

	if len(g.players) < MinPlayersToStart {
		g.message = fmt.Sprintf(MINPLAYERTOSTARTMSG, MinPlayersToStart)
		return
	}

	g.startNewRound()
}

func (g *Game) startNewRound() {

	if len(g.players) == 0 {
		// 플레이어가 없으면 자동으로 로비로 리셋
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
	currentUserName := g.players[randomPlayerIndex].Name
	g.message = fmt.Sprintf(STARTMSG, g.makeNameToDisplay(g.currentUserID, currentUserName))
	log.Printf(STARTLOGMSG, g.RoomId)
}

func (g *Game) eliminatePlayerUnlocked(user *User, reason string) (winner bool, winnerMsg string) {
	for i, p := range g.players {
		if p.ID == user.ID {
			g.players = append(g.players[:i], g.players[i+1:]...)
			g.spectators = append(g.spectators, user)
			g.message = g.makeNameToDisplay(user.ID, user.Name) + ELIMINATEDMSG + reason

			// 현재 차례가 탈락자였으면 다음 활성 플레이어로 이동
			if g.currentUserID == user.ID {
				if len(g.players) > 0 {
					nextIdx := i % len(g.players)
					g.currentUserID = g.players[nextIdx].ID
				} else {
					g.currentUserID = ""
				}
			}

			// 승리 조건: 활성 플레이어가 한 명이면 우승 처리 (잠금은 호출자가 관리)
			if len(g.players) == 1 {
				winner := g.players[0]
				msg := g.makeNameToDisplay(winner.ID, winner.Name) + WINNERMSG
				g.gameover = true
				g.message = msg
				return true, msg
			}

			g.startNewRound()
			return false, ""
		}
	}
	return false, ""
}

func (g *Game) wordDBCheck(word string) bool {
	return g.store.IsWordInDB(word)
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
