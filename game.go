package main

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Game struct {
	room          *Room
	RoomName      string
	RoomId        int
	manager       *RoomManager
	hostUserId    string
	lastWord      string
	usedWords     map[string]bool
	players       []*User
	currentUserID string
	gameover      bool
	started       bool
	message       string
	mu            sync.Mutex
}

type GameMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func NewGame(roomname string, roomId int, manager *RoomManager) *Game {
	//게임 생성시 룸도 같이 생성되게.
	room := NewRoom()
	go room.run()

	return &Game{
		room:      room,
		RoomName:  roomname,
		RoomId:    roomId,
		manager:   manager,
		usedWords: make(map[string]bool),
		players:   make([]*User, 0),
		message:   "플레이어를 기다리는 중...",
		started:   false, //로비상태로 유지.
	}
}

func (g *Game) AddClient(conn *websocket.Conn, name string) {
	user := NewUser(conn, conn.RemoteAddr().String(), name)
	user.game = g

	g.room.register <- user
	g.addUser(user)

	// 새로 접속한 클라이언트에만 개인 welcome 메시지(자기 ID) 전송
	welcome := map[string]string{
		"type":   "welcome",
		"yourId": user.ID,
	}
	if wb, err := json.Marshal(welcome); err == nil {
		if err := user.conn.WriteMessage(websocket.TextMessage, wb); err != nil {
			log.Printf("failed to send welcome to %s: %v", user.ID, err)
		}
	} else {
		log.Println("marshal welcome error:", err)
	}

	// 현재 상태 전파
	g.broadcastGameState()

	// 클라이언트의 메시지 수신 루프 시작 (blocking)
	user.readLoop()

	// readLoop 종료 시 연결 정리
	g.room.unregister <- user
	g.removeUser(user)
	g.broadcastGameState()
}

func (g *Game) handleMessage(user *User, msg []byte) {
	var gameMessage GameMessage

	if err := json.Unmarshal(msg, &gameMessage); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	switch gameMessage.Type {
	case "start_game":
		if !g.started && user.ID == g.hostUserId {
			g.startGame()
			g.broadcastGameState()
		}
	case "submit_word":
		if g.started {
			word, ok := gameMessage.Payload.(string)
			if !ok {
				log.Println("Invalid payload for submit_word:", gameMessage.Payload)
				return
			}
			g.handlePlay(user, word)
		}
	case "reset_game":
		g.mu.Lock()
		g.reset()
		g.mu.Unlock()
		g.broadcastGameState() // reset 후에 상태 전파
	default:
		log.Println("Unknown message type:", gameMessage.Type)
	}

	// g.broadcastGameState()
}

func (g *Game) handlePlay(user *User, word string) {
	g.mu.Lock()

	if g.gameover || g.currentUserID != user.ID {
		g.mu.Unlock()
		return
	}

	word = strings.TrimSpace(word)
	if word == "" {
		g.message = "단어를 입력해주세요."
		g.mu.Unlock()
		g.broadcastGameState() // 메시지만 업데이트하고 상태 전파
		return
	}

	if g.usedWords[word] {
		g.mu.Unlock() // endGame이 락을 관리하므로 먼저 언락
		g.endGame("이미 사용된 단어입니다. 게임 종료!")
		return
	}

	if g.lastWord != "" {
		lastRune, _ := utf8.DecodeLastRuneInString(g.lastWord)
		firstRune, _ := utf8.DecodeRuneInString(word)
		if lastRune != firstRune {
			g.mu.Unlock() // endGame이 락을 관리하므로 먼저 언락
			g.endGame("잘못된 단어입니다! '" + string(lastRune) + "' (으)로 시작해야 합니다. " + "게임 종료!")
			return
		} else if !wordDBCheck(word) {
			g.mu.Unlock() // endGame이 락을 관리하므로 먼저 언락
			g.endGame("사전에 없는 단어입니다! 게임 종료!")
			return
		}
	}

	g.lastWord = word
	g.usedWords[word] = true
	g.setNextPlayerTurn(user.ID)
	g.mu.Unlock()
	g.broadcastGameState() // 다음 턴 상태 전파
}

func (g *Game) endGame(message string) {
	g.mu.Lock()
	g.gameover = true
	g.message = message
	g.mu.Unlock()
	log.Printf("game reset after endGame in room %d", g.RoomId)
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

	g.lastWord = ""
	g.usedWords = make(map[string]bool)
	g.gameover = false
	g.started = false
	g.currentUserID = ""
	g.message = "새 게임을 시작할 수 있습니다. 플레이어를 기다립니다."
	log.Printf("Game reset in room %d", g.RoomId)
}

func (g *Game) startGame() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.started || len(g.players) == 0 {
		return
	}

	g.started = true
	g.gameover = false
	g.currentUserID = g.players[0].ID
	g.message = "게임 시작! " + g.currentUserID + "님부터 시작하세요."
	log.Printf("Game started in room %d", g.RoomId)
}

func (g *Game) broadcastGameState() {
	g.mu.Lock()
	defer g.mu.Unlock()

	players := make([]string, len(g.players))
	for i, player := range g.players {
		players[i] = player.Name
	}

	stateToSend := fiber.Map{
		"lastWord":            g.lastWord,
		"players":             players,
		"currentTurnPlayerId": g.currentUserID,
		"hostUserId":          g.hostUserId,
		"isGameOver":          g.gameover,
		"isStarted":           g.started,
		"message":             g.message,
	}

	bytes, err := json.Marshal(stateToSend)
	if err != nil {
		log.Println("marshal error:", err)
		return
	}

	log.Printf("broadcasting state in room %d: %s", g.RoomId, string(bytes))
	g.room.broadcast <- bytes

}

// 비공개 메서드

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
		log.Printf("Player %s is now the host(ID : %s).", user.Name, user.ID)
		g.reset()
	}
	log.Printf("Player %s Enter the Game(ID : %s)", user.Name, user.ID)
}

func (g *Game) removeUser(user *User) {
	g.mu.Lock()
	shouldDelete := false

	for i, p := range g.players {
		if p.ID == user.ID {
			g.players = append(g.players[:i], g.players[i+1:]...)
			log.Printf("Player %s removed from the game(ID : %s).", user.Name, user.ID)

			if g.hostUserId == user.ID {
				//유저 아무에게 호스트 권한 이전
				if len(g.players) > 0 {
					randomUser := g.players[makeRandomNumber(0, len(g.players))].ID
					g.hostUserId = randomUser
					log.Printf("Host user changed to %s", randomUser)
				} else {
					g.hostUserId = ""
				}
			}

			if g.currentUserID == user.ID && len(g.players) > 0 && !g.gameover {
				nextPlayerIndex := i % len(g.players)
				g.currentUserID = g.players[nextPlayerIndex].ID
				g.message = "플레이어가 나갔습니다. 다음 차례: " + g.currentUserID
			} else if len(g.players) == 0 {
				g.currentUserID = ""
				g.message = "모든 플레이어가 나갔습니다. 새로운 플레이어를 기다립니다."
				g.lastWord = ""
				g.usedWords = make(map[string]bool)
				g.gameover = false
				shouldDelete = true
			}
			g.mu.Unlock()

			// 데드락 가능성 때문에 언락후 삭제.
			if shouldDelete && g.manager != nil {
				log.Printf("Deleting empty room %d", g.RoomId)
				g.manager.DeleteRoom(g.RoomId)
			}
			return
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
			g.message = g.currentUserID + "님의 차례입니다."
			return
		}
	}
}

func wordDBCheck(word string) bool {
	return IsWordInDB(word)
}
