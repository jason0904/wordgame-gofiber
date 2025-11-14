package game

import "log"

func (g *Game) addUser(user *User) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.checkPlayerExist(user)
	g.players = append(g.players, user)

	if len(g.players) == 1 {
		g.handleRoomInit(user)
	}
	log.Printf(ENTERPLAYERLOGMSG, user.Name, user.ID)
}

func (g *Game) removeUser(user *User) {
	g.mu.Lock()
	for i, p := range g.players {
		g.handleDeleteUser(p, user, i)
	}
	for i, s := range g.spectators {
		g.handleDeleteSpectator(s, user, i)
	}
	g.mu.Unlock()
	g.deleteRoom()
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

func (g *Game) handleRoomInit(user *User) {
	g.hostUserId = user.ID
	log.Printf(HOSTLOGMSG, user.Name, user.ID)
	g.reset()
}

func (g *Game) checkPlayerExist(user *User) {
	for _, u := range g.players {
		if u.ID == user.ID {
			return //이미 존재하는 사용자
		}
	}
}

func (g *Game) handleDeleteUser(target, user *User, index int) {

	if target.ID == user.ID {
		g.players = append(g.players[:index], g.players[index+1:]...)
		log.Printf(EXITPLAYERLOGMSG, user.Name, user.ID)
		g.handleHostLeft(user)
		g.makeNewPlayerTurn(user, index)
	}
}

func (g *Game) handleDeleteSpectator(target, user *User, index int) {
	if target.ID == user.ID {
		g.spectators = append(g.spectators[:index], g.spectators[index+1:]...)
		log.Printf(REMOVESPECTATORLOGMSG, user.Name, user.ID)
	}
}

func (g *Game) handleHostLeft(user *User) {
	if g.hostUserId == user.ID {
		g.makeNewHost()
	}
}

func (g *Game) makeNewHost() {
	if len(g.players) > 0 {
		randomUser := g.players[g.random.MakeRandomNumber(0, len(g.players))].ID
		g.hostUserId = randomUser
		log.Printf(HOSTCHANGELOGMSG, randomUser)
	} else {
		g.hostUserId = ""
	}
}

func (g *Game) makeNewPlayerTurn(user *User, index int) {
	if g.currentUserID == user.ID && len(g.players) > 0 && !g.gameover {
		nextPlayerIndex := index % len(g.players)
		g.currentUserID = g.players[nextPlayerIndex].ID
		g.message = EXITMSG + g.currentUserID
	} else if len(g.players) == 0 {
		g.currentUserID = ""
		g.message = ALLEXITMSG
		g.lastWord = ""
		g.usedWords = make(map[string]bool)
		g.gameover = false
	}
}

func (g *Game) deleteRoom() {
	if len(g.players) == 0 {
		log.Printf(DELETEROOMLOGMSG, g.RoomId)
		g.manager.DeleteRoom(g.RoomId)
	}
}
