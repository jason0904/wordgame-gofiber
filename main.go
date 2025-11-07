package main

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func main() {
	app := fiber.New()

	// 정적 파일 서빙
	app.Static("/", "./public")

	roomManager := NewRoomManager()

	// 방 목록 조회 API
	app.Get("/api/rooms", func (c *fiber.Ctx) error {
		rooms := roomManager.GetRooms()
		return c.JSON(rooms)
	})

	app.Post("/api/rooms", func(c *fiber.Ctx) error {
		game := roomManager.MakeRoom("New Room") //나중에 방 이름 설정 할 수 있도록 세팅 변경.
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
