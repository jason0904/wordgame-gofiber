package game

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// ...existing code...
func (g *Game) AddClient(conn *websocket.Conn, name string) {
	id := g.generateUniqueID()
	user := NewUser(conn, id, name)
	user.game = g

	g.room.register <- user
	g.addUser(user)

	// 새로 접속한 클라이언트에만 개인 welcome 메시지(자기 ID) 전송
	welcome := map[string]string{
		"type":   WELCOMEJSONTYPE,
		"yourId": user.ID,
	}
	if wb, err := json.Marshal(welcome); err == nil {
		if err := user.conn.WriteMessage(websocket.TextMessage, wb); err != nil {
			log.Printf(FAILSENDWELCOME, user.ID, err)
		}
	} else {
		log.Println(MARSHALERROR, err)
	}

	// 현재 상태 전파
	g.broadcastGameState()

	// 클라이언트의 메시지 수신 루프 시작 (blocking)
	user.ReadLoop()

	// readLoop 종료 시 연결 정리
	g.room.unregister <- user
	g.removeUser(user)
	g.broadcastGameState()
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
		if g.started {
			word, ok := gameMessage.Payload.(string)
			if !ok {
				log.Println(SUBMITPAYLOADERROR, gameMessage.Payload)
				return
			}
			g.handlePlay(user, word)
		}
	case RESETJSONTYPE:
		g.mu.Lock()
		g.reset()
		g.mu.Unlock()
		g.broadcastGameState() // reset 후에 상태 전파
	default:
		log.Println(UNKNOWNMESSAGETYPE, gameMessage.Type)
	}

}

func (g *Game) broadcastGameState() {
	g.mu.Lock()
	defer g.mu.Unlock()

	players := make([]string, len(g.players))
	for i, player := range g.players {
		players[i] = g.makeNameToDisplay(player.ID, player.Name)
	}

	spectators := make([]string, len(g.spectators))
	for i, s := range g.spectators {
		spectators[i] = g.makeNameToDisplay(s.ID, s.Name)
	}

	stateToSend := fiber.Map{
		"lastWord":            g.lastWord,
		"players":             players,
		"spectators":          spectators,
		"currentTurnPlayerId": g.currentUserID,
		"hostUserId":          g.hostUserId,
		"isGameOver":          g.gameover,
		"isStarted":           g.started,
		"message":             g.message,
	}

	bytes, err := json.Marshal(stateToSend)
	if err != nil {
		log.Println(MARSHALERROR, err)
		return
	}

	log.Printf(BROADCASTLOGMSG, g.RoomId, string(bytes))
	g.room.broadcast <- bytes
}
