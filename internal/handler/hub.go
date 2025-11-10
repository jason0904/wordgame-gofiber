package handler

import (
	"log"
	"strconv"

	"wordgame/internal/game"
	"wordgame/internal/store"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// Handler 는 의존성 주입을 위한 구조체입니다.
type Handler struct {
	RoomManager *game.RoomManager
	DBManager   *store.DBManager
}

func NewHandler(rm *game.RoomManager, db *store.DBManager) *Handler {
	return &Handler{RoomManager: rm, DBManager: db}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/rooms", h.GetRooms)
	api.Post("/rooms", h.CreateRoom)

	app.Get("/ws/:roomId", websocket.New(func(c *websocket.Conn) {
		h.handleWebSocket(c)
	}))
}

func (h *Handler) GetRooms(c *fiber.Ctx) error {
	rooms := h.RoomManager.GetRooms()
	return c.JSON(rooms)
}

type CreateRoomRequest struct {
	RoomName string `json:"roomName"`
}

func (h *Handler) CreateRoom(c *fiber.Ctx) error {
	req := new(CreateRoomRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}
	if req.RoomName == "" {
		req.RoomName = "새로운 방"
	}

	game := h.RoomManager.MakeRoom(req.RoomName, *h.DBManager)
	return c.JSON(fiber.Map{"id": game.RoomId, "roomName": game.RoomName})
}

func (h *Handler) handleWebSocket(conn *websocket.Conn) {
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

	gameObj, exists := h.RoomManager.GetRoom(id)
	if !exists {
		log.Println("Room not found:", roomId)
		_ = conn.Close()
		return
	}

	gameObj.AddClient(conn, name)
}
