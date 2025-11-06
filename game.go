package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Game struct {
	room          *Room
	lastWord      string
	usedWords     map[string]bool
	players       []*User
	currentUserID string
	gameover      bool
	message       string
	mu            sync.Mutex
}

type GameMessage struct {
	Type    string `json:"type"`
	Payload string `json:"data"`
}

// 전역 난수 생성기: rand.Seed 대신 rand.New(rand.NewSource(...)) 사용
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (g *Game) NewGame(room *Room) *Game {

	return &Game{
		room:      room,
		usedWords: make(map[string]bool),
		players:   make([]*User, 0),
		message:   "플레이어를 기다리는 중...",
	}
}

func (g *Game) AddClient(conn *websocket.Conn) {
	user := &User{
		conn: conn,
		ID: conn.RemoteAddr().String(),
		game: g,
	}

	g.room.register <- user
	g.addUser(user)

	if len(g.players) == 1 {
		g.reset()
	}

	g.broadcastGameState()

	//클라이언트의 메시지 수신 루프 시작
	user.readLoop()

	//클라이언트 연결 종료 처리
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
	case "word":
		g.handlePlay(user, gameMessage.Payload)
	case "reset_game":
		g.reset()
	default:
		log.Println("Unknown message type:", gameMessage.Type)
	}

	g.broadcastGameState()
}

func (g *Game) handlePlay(user *User, word string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.gameover || g.currentUserID != user.ID {
		return
	}

	word = strings.TrimSpace(word)
	if word == "" {
		g.message = "단어를 입력해주세요."
		return
	}

	if g.usedWords[word] {
		g.gameover = true
		g.message = "이미 사용된 단어입니다. 게임 종료!"
		return
	}

	if g.lastWord != "" {
		lastRune, _ := utf8.DecodeLastRuneInString(g.lastWord)
		firstRune, _ := utf8.DecodeRuneInString(word)
		if lastRune != firstRune {
			g.gameover = true
			g.message = "잘못된 단어입니다! '" + string(lastRune) + "' (으)로 시작해야 합니다. " + "게임 종료!"
			return
		}
	}

	g.lastWord = word
	g.usedWords[word] = true
	g.setNextPlayerTurn(user.ID)
}

func (g *Game) reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.lastWord = ""
	g.usedWords = make(map[string]bool)
	g.gameover = false

	if len(g.players) > 0 {
		// 시작 플레이어를 무작위로 선택
		idx := rnd.Intn(len(g.players))
		g.currentUserID = g.players[idx].ID
		g.message = "게임 시작! " + g.currentUserID + "님의 차례입니다."
	} else {
		g.currentUserID = ""
		g.message = "플레이어를 기다리는 중..."
	}
	log.Println("game reset")
}

func (g *Game) broadcastGameState() {
	g.mu.Lock()
	defer g.mu.Unlock()

	players := make([]string, len(g.players))
	for i, player := range g.players {
		players[i] = player.ID
	}

	stateToSend := fiber.Map{
		"lastWord":            g.lastWord,
		"players":             players,
		"currentTurnPlayerId": g.currentUserID,
		"isGameOver":          g.gameover,
		"message":             g.message,
	}

	bytes, err := json.Marshal(stateToSend)
	if err != nil {
		log.Println("marshal error:", err)
		return
	}

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
	log.Printf("Player %s Enter the Game", user.ID)
}

func (g *Game) removeUser(user *User) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, p := range g.players {
		if p.ID == user.ID {
			g.players = append(g.players[:i], g.players[i+1:]...)
			log.Printf("Player %s removed from the game.", user.ID)

			if g.currentUserID == user.ID && len(g.players) > 0 && !g.gameover {
				nextPlayerIndex := i % len(g.players)
				g.currentUserID = g.players[nextPlayerIndex].ID
				g.message = "플레이어가 나갔습니다. 다음 차례: " + g.currentUserID
			} else if len(g.players) == 0 {
				//reset 호출하는 대신 게임 상태 초기화
				g.currentUserID = ""
				g.message = "모든 플레이어가 나갔습니다. 새로운 플레이어를 기다립니다."
				g.lastWord = ""
				g.usedWords = make(map[string]bool)
				g.gameover = false
			}
			return
		}
	}
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


