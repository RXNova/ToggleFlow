package main

import (
	"log"
	"os"

	"github.com/RXNova/ToggleFlow/internal/api"
	"github.com/RXNova/ToggleFlow/internal/db"
	"github.com/RXNova/ToggleFlow/internal/ui"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "ToggleFlow",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	// API routes — like registering a NestJS module
	api.Register(app, database)

	// Serve embedded Vue dashboard — must be registered last
	ui.Register(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
