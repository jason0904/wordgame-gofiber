package game

import (
	"log"
	"sync"

	"wordgame/internal/random"
	"wordgame/internal/store"
)

type RoomManager struct {
	rooms map[int]*Game
	random *random.Manager

	mutex sync.RWMutex
}

func NewRoomManager(random *random.Manager) *RoomManager {
	return &RoomManager{
		rooms: make(map[int]*Game),
		random: random,
	}
}

func (rm *RoomManager) MakeRoom(name string, db *store.DBManager) *Game {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	roomId := rm.generateRoomID()
	room := NewGame(name, roomId, rm, rm.random, db)
	rm.rooms[roomId] = room
	log.Printf("Room created: %d", roomId)
	return room
}

func (rm *RoomManager) GetRoom(id int) (*Game, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	room, exists := rm.rooms[id]
	return room, exists
}

func (rm *RoomManager) GetRooms() []map[string]any {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	list := make([]map[string]any, 0, len(rm.rooms))
	for id, game := range rm.rooms {
		list = append(list, map[string]any{
			"id":          id,
			"roomName":    game.RoomName,
			"playerCount": len(game.players),
			"isStarted":   game.started,
		})
	}
	return list
}

func (rm *RoomManager) DeleteRoom(id int) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	delete(rm.rooms, id)
}

func (rm *RoomManager) generateRoomID() int {
	return rm.random.MakeRandomNumber(MINROOMIDIDENTIFIER, MAXROOMIDIDENTIFIER)
}
