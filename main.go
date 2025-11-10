package main

import (
	"log"

	"wordgame/internal/game"
	"wordgame/internal/handler"
	"wordgame/internal/random"
	"wordgame/internal/store"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Static("/", "./assets/public")

	randomManager := random.NewManager()
	roomManager := game.NewRoomManager(randomManager)
	dbManager, err := store.NewDBManager()
	if err != nil {
		log.Fatalf("Failed to initialize database manager: %v", err)
	}

	apiHandler := handler.NewAPIHandler(roomManager, dbManager)
	apiHandler.RegisterRoutes(app)

	wsHandler := handler.NewWSHandler(roomManager)
	wsHandler.RegisterRoutes(app)

	log.Println("listening on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
