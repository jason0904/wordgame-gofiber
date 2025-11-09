package game

import (
	"sync"
	"wordgame/internal/random"
	"wordgame/internal/store"

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
	store         store.DBManager
	random        random.Manager
}

type GameMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func NewGame(roomname string, roomId int, manager *RoomManager, rnd random.Manager, store store.DBManager) *Game {
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
		message:    WAITINGFORPLAYERSMSG,
		startword:  "",
		started:    false, //로비상태로 유지.
		random:     rnd,
		store:      store,
	}
}