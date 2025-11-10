package game

import (
	"os"
	"github.com/joho/godotenv"
	"wordgame/internal/random"
	"wordgame/internal/store"
)

var testGame *Game

func SetupTestGame() *Game {
	if err := os.Chdir("../../"); err != nil {
		panic("could not change to root dir")
	}

	if err := godotenv.Load(); err != nil {
		// .env 파일이 없어도 패닉을 발생시키지 않고 진행할 수 있습니다.
		// NewDBManager가 기본 경로를 사용하게 됩니다.
	}
	if testGame == nil {
		dbManager, err := store.NewDBManager()
		if err != nil {
			panic(err)
		}
		rm := NewRoomManager(*random.NewManager())
		testGame = NewGame("Test Room", 1, rm, *random.NewManager(), *dbManager)
	}

	return testGame
}

func SetupDefaultPlayers() *Game {
	g := SetupTestGame()
	g.mu.Lock()
	defer g.mu.Unlock()
	g.players = []*User{
		{ID: "1001", Name: "Alice"},
		{ID: "1002", Name: "Bob"},
		{ID: "1003", Name: "Charlie"},
	}
	g.currentUserID = "1001"
	g.hostUserId = "1001"
	return g
}
