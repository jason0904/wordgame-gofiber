package main

import (
	"log"

	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// 정적 파일 서빙
	app.Static("/", "./public")

	roomManager := NewRoomManager()

	// 방 목록 조회 API
	app.Get("/api/rooms", func(c *fiber.Ctx) error {
		rooms := roomManager.GetRooms()
		return c.JSON(rooms)
	})

	// 방 생성 요청을 위한 구조체
	type CreateRoomRequest struct {
		RoomName string `json:"roomName"`
	}

	app.Post("/api/rooms", func(c *fiber.Ctx) error {
		req := new(CreateRoomRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
		}

		if req.RoomName == "" {
			req.RoomName = "새로운 방" // 이름이 없으면 기본값 사용
		}

		game := roomManager.MakeRoom(req.RoomName)
		return c.JSON(fiber.Map{"id": game.RoomId, "roomName": game.RoomName})
	})

	//websocket 핸들러
	app.Get("/ws/:roomId", websocket.New(func(c *websocket.Conn) {
		roomId := c.Params("roomId")

		name := c.Query("name")

		if name == "" {
			name = "user"
		}

		id, err := strconv.Atoi(roomId)
		if err != nil {
			log.Println("Invalid room ID:", roomId)
			return
		}

		game, exists := roomManager.GetRoom(id)

		if !exists {
			log.Println("Room not found:", roomId)
			return
		}

		game.AddClient(c, name)
	}))

	log.Println("listening on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
