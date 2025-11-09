package game

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"wordgame/internal/store"
	"wordgame/internal/random"
)

const (
	MinPlayersToStart  = 2
	MinStartWordLength = 2
	MaxStartWordLength = 6
)

type Game struct {
	room          *Room
	RoomName      string
	RoomId        int
	manager       *RoomManager
	hostUserId    string
	lastWord      string
	startword     string
	usedWords     map[string]bool
	players       []*User
	spectators    []*User
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
	go room.Run()

	return &Game{
		room:       room,
		RoomName:   roomname,
		RoomId:     roomId,
		manager:    manager,
		usedWords:  make(map[string]bool),
		players:    make([]*User, 0),
		spectators: make([]*User, 0),
		message:    "플레이어를 기다리는 중...",
		startword:  "",
		started:    false, //로비상태로 유지.
	}
}

func (g *Game) AddClient(conn *websocket.Conn, name string) {
	id := g.generateUniqueID()
	user := NewUser(conn, id, name)
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
	user.ReadLoop()

	// readLoop 종료 시 연결 정리
	g.room.unregister <- user	
	g.removeUser(user)
	g.broadcastGameState()
}

func (g *Game) HandleMessage(user *User, msg []byte) {
	var gameMessage GameMessage

	if err := json.Unmarshal(msg, &gameMessage); err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	switch gameMessage.Type {
	case "start_game":
		g.startGame(user)
		g.broadcastGameState()
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

}

// 비공개 메서드

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
		g.broadcastGameState()
		return
	}

	if utf8.RuneCountInString(word) < 2 {
		g.message = "단어는 최소 2자 이상이어야 합니다."
		g.mu.Unlock()
		g.broadcastGameState()
		return
	}

	if g.usedWords[word] {
		winner, msg := g.eliminatePlayerUnlocked(user, "이미 사용된 단어입니다.")
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
			winner, msg := g.eliminatePlayerUnlocked(user, "끝말이 맞지 않습니다.")
			g.mu.Unlock()
			if winner {
				g.endGame(msg)
			} else {
				g.broadcastGameState()
			}
			return
		} else if !wordDBCheck(word) {
			winner, msg := g.eliminatePlayerUnlocked(user, "사전에 없는 단어입니다.")
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
	g.message = "새 게임을 시작할 수 있습니다. 플레이어를 기다립니다."
	log.Printf("Game reset in room %d", g.RoomId)
}

func (g *Game) startGame(user *User) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.started {
		g.message = "이미 게임이 시작되었습니다."
		return
	}

	if g.hostUserId != user.ID {
		g.message = "게임을 시작할 권한이 없습니다. 호스트만 게임을 시작할 수 있습니다."
		return
	}

	if len(g.players) < MinPlayersToStart {
		g.message = "게임을 시작하려면 최소 " + strconv.Itoa(MinPlayersToStart) + "명의 플레이어가 필요합니다."
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
	g.message = "게임 시작! " + g.makeNameToDisplay(g.currentUserID, currentUserName) + "님부터 시작하세요."
	log.Printf("Game started in room %d", g.RoomId)
}

func (g *Game) eliminatePlayerUnlocked(user *User, reason string) (winner bool, winnerMsg string) {
	for i, p := range g.players {
		if p.ID == user.ID {
			g.players = append(g.players[:i], g.players[i+1:]...)
			g.spectators = append(g.spectators, user)
			g.message = g.makeNameToDisplay(user.ID, user.Name) + "님이 탈락했습니다. 이유: " + reason

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
				msg := g.makeNameToDisplay(winner.ID, winner.Name) + "님이 최종 우승했습니다!"
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
		log.Println("marshal error:", err)
		return
	}

	log.Printf("broadcasting state in room %d: %s", g.RoomId, string(bytes))
	g.room.broadcast <- bytes
}


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
					randomUser := g.players[random.MakeRandomNumber(0, len(g.players))].ID
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

	for i, s := range g.spectators {
		if s.ID == user.ID {
			g.spectators = append(g.spectators[:i], g.spectators[i+1:]...)
			log.Printf("Spectator %s removed(ID : %s).", user.Name, user.ID)
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
			g.message = g.makeNameToDisplay(g.currentUserID, nextPlayerName) + "님의 차례입니다."
			return
		}
	}
}

func wordDBCheck(word string) bool {
	return store.IsWordInDB(word)
}

func (g *Game) generateUniqueID() string {
	const maxAttempts = 10000
	for i := 0; i < maxAttempts; i++ {
		n := random.MakeRandomNumber(1000, 10000)
		id := strconv.Itoa(n)

		g.mu.Lock()
		exists := false
		for _, p := range g.players {
			if p.ID == id {
				exists = true
				break
			}
		}
		if !exists {
			for _, s := range g.spectators {
				if s.ID == id {
					exists = true
					break
				}
			}
		}
		g.mu.Unlock()

		if !exists {
			return id
		}
	}
	log.Println("Warning: generateUniqueID reached max attempts, returning fallback ID")
	return "0000"
}

func (g *Game) selectRandomPlayerIndex() int {
	return random.MakeRandomNumber(0, len(g.players))
}

func (g *Game) makeStartWord() string {
	randomWordLength := random.MakeRandomNumber(MinStartWordLength, MaxStartWordLength) // 2자에서 6자 사이
	word, err := store.GetRandomWordByLength(randomWordLength)
	if err != nil {
		log.Println("Error getting random start word:", err)
		return "사과" // 기본 단어 반환
	}
	return word
}

func (g *Game) makeNameToDisplay(id string, name string) string {
	return name + "#" + id
}
