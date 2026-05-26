package api

import (
	"toggleflow/internal/auth"
	"toggleflow/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
)

func Register(app *fiber.App, database *bun.DB) {
	h := newHandler(database)

	// Public — no auth required
	app.Get("/api/setup/status", h.SetupStatus)
	app.Post("/api/setup", h.Setup)
	app.Post("/api/auth/login", h.Login)

	// Authenticated — all roles
	protected := app.Group("/api", auth.Require)
	protected.Get("/auth/me", h.Me)
	protected.Patch("/auth/profile", h.UpdateProfile)

	// User management — Admin and above
	users := protected.Group("/users", auth.RequireRole(db.RoleAdmin))
	users.Get("/", h.ListUsers)
	users.Post("/", h.CreateUser)
	users.Patch("/:id", h.UpdateUser)

	// Delete user — Superuser only
	protected.Delete("/users/:id", auth.RequireRole(db.RoleSuperuser), h.DeleteUser)

	// Projects — Owner and above
	projects := protected.Group("/projects", auth.RequireRole(db.RoleOwner))
	projects.Post("/", h.CreateProject)
	projects.Get("/", h.ListProjects)
	projects.Patch("/:id", h.UpdateProject)
	projects.Delete("/:id", h.DeleteProject)
	projects.Post("/:pid/environments", h.CreateEnvironment)
	projects.Get("/:pid/environments", h.ListEnvironments)

	// Flags — Editor and above to write, Viewer to read
	protected.Get("/projects/:pid/flags", h.ListFlags)
	protected.Get("/projects/:pid/flags/:key", h.GetFlag)
	protected.Get("/projects/:pid/audit", h.ListAudit)

	flagsWrite := protected.Group("/projects/:pid/flags", auth.RequireRole(db.RoleEditor))
	flagsWrite.Post("/", h.CreateFlag)
	flagsWrite.Patch("/:key", h.UpdateFlag)
	flagsWrite.Delete("/:key", h.DeleteFlag)

	// SDK — authenticated by sdk_key query param, not JWT
	sdk := app.Group("/sdk")
	sdk.Get("/flags", h.SDKGetFlags)
	sdk.Post("/evaluate", h.SDKEvaluate)
	sdk.Get("/stream", h.SDKStream)
}
