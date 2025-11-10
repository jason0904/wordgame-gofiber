package game

import (
	"wordgame/internal/random"
	"wordgame/internal/store"

	"testing"
	"github.com/stretchr/testify/assert"
)


func TestMakeRoom(t *testing.T) {
	randomManager := random.NewManager()
	rm := NewRoomManager(randomManager)
	dbMock := &store.DBManager{}
	
	roomName := "Test Room"
	room := rm.MakeRoom(roomName, dbMock)
	assert.NotNil(t, room, "MakeRoom should return a non-nil room")
	assert.Equal(t, roomName, room.RoomName, "Room name should match the provided name")

	retrievedRoom, exists := rm.GetRoom(room.RoomId)
	assert.True(t, exists, "GetRoom should find the created room")
	assert.Equal(t, room, retrievedRoom, "Retrieved room should match the created room")
}

func TestGetRooms(t *testing.T) {
	randomManager := random.NewManager()
	rm := NewRoomManager(randomManager)
	dbMock := &store.DBManager{}
	
	room1 := rm.MakeRoom("Room 1", dbMock)
	room2 := rm.MakeRoom("Room 2", dbMock)
	rooms := rm.GetRooms()
	assert.Len(t, rooms, 2, "There should be 2 rooms in the manager")

	for _, r := range rooms {
		id := r["id"].(int)
		if id == room1.RoomId {
			assert.Equal(t, "Room 1", r["roomName"], "Room name should match for Room 1")
		} else if id == room2.RoomId {
			assert.Equal(t, "Room 2", r["roomName"], "Room name should match for Room 2")
		} else {
			t.Errorf("Unexpected room ID: %d", id)
		}
	}
}

func TestDeleteRoom(t *testing.T) {
	randomManager := random.NewManager()
	rm := NewRoomManager(randomManager)
	dbMock := &store.DBManager{}

	room := rm.MakeRoom("Room to Delete", dbMock)
	rm.DeleteRoom(room.RoomId)
	
	_, exists := rm.GetRoom(room.RoomId)
	assert.False(t, exists, "Room should not exist after deletion")
}