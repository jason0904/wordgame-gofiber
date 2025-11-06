package main

import (
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
	room := NewGame(name)
	rm.rooms[roomId] = room
	return room
}

func (rm *RoomManager) GetRoom(id int) (*Game, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	room, exists := rm.rooms[id]
	return room, exists
}

func (rm *RoomManager) GetRooms() map[int]*Game {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	return rm.rooms
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