package handler

import (
	"wordgame/internal/game"
	"wordgame/internal/store"

	"github.com/gofiber/fiber/v2"
)

type APIHandler struct {
	RoomManager *game.RoomManager
	DBManager   *store.DBManager
}

func NewAPIHandler(rm *game.RoomManager, db *store.DBManager) *APIHandler {
	return &APIHandler{RoomManager: rm, DBManager: db}
}

func (a *APIHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Get("/rooms", a.GetRooms)
	api.Post("/rooms", a.CreateRoom)

}

func (a *APIHandler) GetRooms(c *fiber.Ctx) error {
	rooms := a.RoomManager.GetRooms()
	return c.JSON(rooms)
}

type CreateRoomRequest struct {
	RoomName string `json:"roomName"`
}

func (a *APIHandler) CreateRoom(c *fiber.Ctx) error {
	req := new(CreateRoomRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}
	if req.RoomName == "" {
		req.RoomName = "새로운 방"
	}

	game := a.RoomManager.MakeRoom(req.RoomName, a.DBManager)
	return c.JSON(fiber.Map{"id": game.RoomId, "roomName": game.RoomName})
}
