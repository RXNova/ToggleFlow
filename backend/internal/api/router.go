package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
)

// Register wires all routes onto the Fiber app.
// Think of this as the root AppModule in NestJS — it pulls in all sub-modules.
func Register(app *fiber.App, db *bun.DB) {
	h := newHandler(db)

	// Health check
	app.Get("/api/health", h.Health)

	// Projects — like a NestJS controller at /projects
	projects := app.Group("/api/projects")
	projects.Post("/", h.CreateProject)
	projects.Get("/", h.ListProjects)

	// Environments — nested under a project
	projects.Post("/:pid/environments", h.CreateEnvironment)
	projects.Get("/:pid/environments", h.ListEnvironments)

	// Flags
	projects.Post("/:pid/flags", h.CreateFlag)
	projects.Get("/:pid/flags", h.ListFlags)
	projects.Get("/:pid/flags/:key", h.GetFlag)
	projects.Patch("/:pid/flags/:key", h.UpdateFlag)
	projects.Delete("/:pid/flags/:key", h.DeleteFlag)

	// Audit log
	projects.Get("/:pid/audit", h.ListAudit)

	// SDK endpoints — authenticated by ?sdk_key= query param
	sdk := app.Group("/sdk")
	sdk.Get("/flags", h.SDKGetFlags)
	sdk.Post("/evaluate", h.SDKEvaluate)
	sdk.Get("/stream", h.SDKStream)
}
