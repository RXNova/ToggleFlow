package api

import (
	"toggleflow/internal/auth"
	"toggleflow/internal/db"
	"toggleflow/internal/stream"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
)

func Register(app *fiber.App, database *bun.DB, broker *stream.Broker) {
	h := newHandler(database, broker)

	// Public — no auth required
	app.Get("/api/setup/status", h.SetupStatus)
	app.Post("/api/setup", h.Setup)
	app.Post("/api/auth/login", h.Login)
	app.Post("/api/auth/activate", h.ActivateAccount)
	app.Get("/api/auth/invite/:uuid", h.GetInviteInfo)
	app.Post("/api/auth/reset", h.ResetPassword)
	app.Get("/api/auth/reset/:uuid", h.GetResetInfo)

	// Authenticated — JWT or project-scoped API key
	protected := app.Group("/api", h.requireAuth)
	protected.Get("/auth/me", h.Me)
	protected.Patch("/auth/profile", h.UpdateProfile)

	// User management — Admin and above (JWT only, API keys can't manage users)
	users := protected.Group("/users", auth.RequireRole(db.RoleAdmin))
	users.Get("/", h.ListUsers)
	users.Post("/", h.CreateUser)
	users.Patch("/:id", h.UpdateUser)
	users.Post("/:id/reinvite", h.ReinviteUser)
	users.Get("/:id/audit", h.ListUserAudit)

	// Delete user and reset link — Superuser only
	protected.Delete("/users/:id", auth.RequireRole(db.RoleSuperuser), h.DeleteUser)
	protected.Post("/users/:id/reset-link", auth.RequireRole(db.RoleSuperuser), h.GenerateResetLink)

	// Projects + environments list — any authenticated member (membership gate inside handlers)
	protected.Get("/projects", h.ListProjects)
	protected.Get("/projects/:pid/environments", h.ListEnvironments)
	protected.Get("/projects/:pid/members", h.ListMembers)

	// Flags — Editor and above to write, any member to read
	protected.Get("/projects/:pid/flags", h.ListFlags)
	protected.Get("/projects/:pid/flags/:key", h.GetFlag)
	protected.Get("/projects/:pid/audit", h.ListAudit)

	// Segments — any member to read, Editor and above to write
	protected.Get("/projects/:pid/segments", h.ListSegments)

	flagsWrite := protected.Group("/projects/:pid/flags", auth.RequireRole(db.RoleEditor))
	flagsWrite.Post("/", h.CreateFlag)
	flagsWrite.Patch("/:key", h.UpdateFlag)
	flagsWrite.Patch("/:key/env", h.ToggleFlagEnv)
	flagsWrite.Put("/:key/rules", h.SaveFlagRules)
	flagsWrite.Delete("/:key", h.DeleteFlag)

	segmentsWrite := protected.Group("/projects/:pid/segments", auth.RequireRole(db.RoleEditor))
	segmentsWrite.Post("/", h.CreateSegment)
	segmentsWrite.Patch("/:sid", h.UpdateSegment)
	segmentsWrite.Delete("/:sid", h.DeleteSegment)

	// Project + environment write ops — Owner and above
	projectsOwner := protected.Group("/projects", auth.RequireRole(db.RoleOwner))
	projectsOwner.Post("/", h.CreateProject)
	projectsOwner.Patch("/:id", h.UpdateProject)
	projectsOwner.Delete("/:id", h.DeleteProject)
	projectsOwner.Post("/:pid/environments", h.CreateEnvironment)
	projectsOwner.Patch("/:pid/environments/:eid", h.UpdateEnvironment)
	projectsOwner.Delete("/:pid/environments/:eid", h.DeleteEnvironment)
	projectsOwner.Get("/:pid/environments/:eid/sdk-keys", h.ListSDKKeys)
	projectsOwner.Post("/:pid/environments/:eid/sdk-keys", h.CreateSDKKey)
	projectsOwner.Delete("/:pid/environments/:eid/sdk-keys/:kid", h.DeleteSDKKey)
	projectsOwner.Get("/:pid/api-keys", h.ListAPIKeys)
	projectsOwner.Post("/:pid/api-keys", h.CreateAPIKey)
	projectsOwner.Delete("/:pid/api-keys/:kid", h.DeleteAPIKey)
	projectsOwner.Post("/:pid/members", h.AddMember)
	projectsOwner.Delete("/:pid/members/:uid", h.RemoveMember)

	// SDK — authenticated by sdk_key query param, not JWT
	sdk := app.Group("/sdk")
	sdk.Get("/flags", h.SDKGetFlags)
	sdk.Post("/evaluate", h.SDKEvaluate)
	sdk.Get("/stream", h.SDKStream)
}
