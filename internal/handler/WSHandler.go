package handler

import (
	"log"
	"strconv"
	
	"wordgame/internal/game"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

)

type WSHandler struct {
	RoomManager *game.RoomManager
}

func NewWSHandler(rm *game.RoomManager) *WSHandler {
	return &WSHandler{RoomManager: rm}
}


func (ws *WSHandler) RegisterRoutes(app fiber.Router) {
	app.Get("/ws/:roomId", websocket.New(func(c *websocket.Conn) {
		ws.handleWebSocket(c)
	}))
}

func (ws *WSHandler) handleWebSocket(conn *websocket.Conn) {
	roomId := conn.Params("roomId")
	name := conn.Query("name")
	if name == "" {
		name = "user"
	}

	id, err := strconv.Atoi(roomId)
	if err != nil {
		log.Println("Invalid room ID:", roomId)
		_ = conn.Close()
		return
	}

	gameObj, exists := ws.RoomManager.GetRoom(id)
	if !exists {
		log.Println("Room not found:", roomId)
		_ = conn.Close()
		return
	}

	gameObj.AddClient(conn, name)
}
