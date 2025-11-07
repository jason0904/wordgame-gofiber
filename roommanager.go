package main

import (
	"log"
	"sync"
)

type RoomManager struct {
	rooms map[int]*Game

	mutex sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[int]*Game),
	}
}

func (rm *RoomManager) MakeRoom(name string) *Game {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	roomId := generateRoomID()
	room := NewGame(name, roomId, rm)
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
			"id":           id,
			"room_name":     game.RoomName,
			"player_count": len(game.players),
			"is_started":   game.started,
		})
	}
	return list
}

func (rm *RoomManager) DeleteRoom(id int) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	delete(rm.rooms, id)
}

// 비공개 메서드

func generateRoomID() int {
	return makeRandomNumber(1000, 9000)
}
