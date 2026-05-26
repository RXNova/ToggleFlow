package ui

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// Vue dist/ is embedded into the binary at compile time via the directive below.
// In dev, this folder is empty — Vite serves the frontend at :5173 instead.
//
//go:embed dist
var dist embed.FS

// Register mounts the Vue SPA on all non-API routes.
// NotFoundFile ensures deep links (e.g. /flags/my-flag) fall back to index.html — like Angular's pathMatch: 'full'.
func Register(app *fiber.App) {
	sub, _ := fs.Sub(dist, "dist")
	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(sub),
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))
}
