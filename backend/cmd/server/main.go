package main

import (
	"log"
	"os"
	"time"

	"toggleflow/internal/api"
	"toggleflow/internal/db"
	"toggleflow/internal/stream"
	"toggleflow/internal/ui"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() { _ = database.Close() }()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if os.Getenv("SEED_DEMO") == "true" {
		if err := db.Seed(database); err != nil {
			log.Fatalf("failed to seed demo data: %v", err)
		}
	}

	app := fiber.New(fiber.Config{
		AppName: "ToggleFlow",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	broker := stream.New()

	if os.Getenv("SEED_DEMO") == "true" {
		go func() {
			ticker := time.NewTicker(30 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				if err := db.Reset(database); err != nil {
					log.Printf("demo reset failed: %v", err)
					continue
				}
				if err := db.Seed(database); err != nil {
					log.Printf("demo re-seed failed: %v", err)
					continue
				}
				broker.PublishAll(stream.Event{Action: "demo.reset"})
				log.Println("demo database reset and re-seeded")
			}
		}()
	}

	// API routes — like registering a NestJS module
	api.Register(app, database, broker)

	// Serve embedded Vue dashboard — must be registered last
	ui.Register(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
