package game

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type WelcomeMessage struct {
	Type   string `json:"type"`
	YourId string `json:"yourId"`
}

func (g *Game) AddClient(conn *websocket.Conn, name string) {
	id := g.generateUniqueID()
	user := NewUser(conn, id, name)
	user.game = g

	g.room.register <- user
	g.addUser(user)
	welcome := g.makeWelcomeMessage(user)

	if welcomeJson, err := json.Marshal(welcome); err == nil {
		g.sendMessageToUser(user, welcomeJson)
	} else {
		log.Println(MARSHALERROR, err)
	}

	g.handleAfterConnect(user)
}

func (g *Game) HandleMessage(user *User, msg []byte) {
	var gameMessage GameMessage

	if err := json.Unmarshal(msg, &gameMessage); err != nil {
		log.Println(UNMARSHALERROR, err)
		return
	}

	switch gameMessage.Type {
	case STARTJSONTYPE:
		g.startGame(user)
		g.broadcastGameState()
	case SUBMITJSONTYPE:
		g.handleSubmit(user, gameMessage)
	default:
		log.Println(UNKNOWNMESSAGETYPE, gameMessage.Type)
	}
}

func (g *Game) broadcastGameState() {
	g.mu.Lock()
	defer g.mu.Unlock()

	players := g.makePlayerList()
	spectators := g.makeSpectatorList()
	stateToSend := g.makeSendForm(players, spectators)
	bytes, err := json.Marshal(stateToSend)
	if err != nil {
		log.Println(MARSHALERROR, err)
		return
	}
	log.Printf(BROADCASTLOGMSG, g.RoomId, string(bytes))
	g.room.broadcast <- bytes
}

func (g *Game) handleSubmit(user *User, gameMessage GameMessage) {
	if !g.started {
		return
	}
	word, ok := gameMessage.Payload.(string)
	if !ok {
		log.Println(SUBMITPAYLOADERROR, gameMessage.Payload)
		return
	}
	g.handlePlay(user, word)
}

func (g *Game) makeWelcomeMessage(user *User) WelcomeMessage {
	return WelcomeMessage{
		Type:   "welcome",
		YourId: user.ID,
	}
}

func (g *Game) handleAfterConnect(user *User) {
	g.broadcastGameState()
	user.ReadLoop()
	g.handleClientDisconnect(user)
}

func (g *Game) handleClientDisconnect(user *User) {
	user.Close()
	g.room.unregister <- user
	g.removeUser(user)
	g.broadcastGameState()
}

func (g *Game) sendMessageToUser(user *User, json []byte) {
	if err := user.WriteMessage(websocket.TextMessage, json); err != nil {
		log.Printf(FAILSENDWELCOME, user.ID, err)
	}
}

func (g *Game) makePlayerList() []string {
	players := make([]string, len(g.players))
	for i, player := range g.players {
		players[i] = g.makeNameToDisplay(player.ID, player.Name)
	}
	return players
}

func (g *Game) makeSpectatorList() []string {
	spectators := make([]string, len(g.spectators))
	for i, spectator := range g.spectators {
		spectators[i] = g.makeNameToDisplay(spectator.ID, spectator.Name)
	}
	return spectators
}

func (g *Game) makeSendForm(players, spectators []string) fiber.Map {
	return fiber.Map{
		"lastWord":            g.lastWord,
		"players":             players,
		"spectators":          spectators,
		"currentTurnPlayerId": g.currentUserID,
		"hostUserId":          g.hostUserId,
		"isGameOver":          g.gameover,
		"isStarted":           g.started,
		"message":             g.message,
	}
}
