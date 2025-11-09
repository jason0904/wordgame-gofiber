package game

import "log"

// ...existing code...
func (g *Game) addUser(user *User) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, u := range g.players {
		if u.ID == user.ID {
			return //이미 존재하는 사용자
		}
	}
	g.players = append(g.players, user)

	if len(g.players) == 1 {
		g.hostUserId = user.ID
		log.Printf(HOSTLOGMSG, user.Name, user.ID)
		g.reset()
	}
	log.Printf(ENTERPLAYERLOGMSG, user.Name, user.ID)
}

func (g *Game) removeUser(user *User) {
	g.mu.Lock()
	shouldDelete := false

	for i, p := range g.players {
		if p.ID == user.ID {
			g.players = append(g.players[:i], g.players[i+1:]...)
			log.Printf(EXITPLAYERLOGMSG, user.Name, user.ID)

			if g.hostUserId == user.ID {
				//유저 아무에게 호스트 권한 이전
				if len(g.players) > 0 {
					randomUser := g.players[g.random.MakeRandomNumber(0, len(g.players))].ID
					g.hostUserId = randomUser
					log.Printf(HOSTCHANGELOGMSG, randomUser)
				} else {
					g.hostUserId = ""
				}
			}

			if g.currentUserID == user.ID && len(g.players) > 0 && !g.gameover {
				nextPlayerIndex := i % len(g.players)
				g.currentUserID = g.players[nextPlayerIndex].ID
				g.message = EXITMSG + g.currentUserID
			} else if len(g.players) == 0 {
				g.currentUserID = ""
				g.message = ALLEXITMSG
				g.lastWord = ""
				g.usedWords = make(map[string]bool)
				g.gameover = false
				shouldDelete = true
			}
			g.mu.Unlock()

			// 데드락 가능성 때문에 언락후 삭제.
			if shouldDelete && g.manager != nil {
				log.Printf(DELETEROOMLOGMSG, g.RoomId)
				g.manager.DeleteRoom(g.RoomId)
			}
			return
		}
	}

	for i, s := range g.spectators {
		if s.ID == user.ID {
			g.spectators = append(g.spectators[:i], g.spectators[i+1:]...)
			log.Printf(REMOVESPECTATORLOGMSG, user.Name, user.ID)
			break
		}
	}
	g.mu.Unlock()
}

func (g *Game) setNextPlayerTurn(currentUserID string) {
	if len(g.players) == 0 {
		g.currentUserID = ""
		return
	}

	for i, p := range g.players {
		if p.ID == currentUserID {
			nextPlayerIndex := (i + 1) % len(g.players)
			g.currentUserID = g.players[nextPlayerIndex].ID
			nextPlayerName := g.players[nextPlayerIndex].Name
			g.message = g.makeNameToDisplay(g.currentUserID, nextPlayerName) + CURRENTTURNMSG
			return
		}
	}
}
