package api

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"

	"toggleflow/internal/auth"
	"toggleflow/internal/db"
	"toggleflow/internal/stream"
)

const apiKeyProjectLocal = "api_key_project_id"

// Page is the standard paginated response shape for all list endpoints.
// Using Go generics (available since 1.18) so every handler returns the same
// envelope without losing type info — similar to a generic DTO wrapper in NestJS.
type Page[T any] struct {
	Data   []T `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type pageQuery struct {
	Limit  int
	Offset int
	Search string
}

func parsePage(c *fiber.Ctx) pageQuery {
	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}
	return pageQuery{Limit: limit, Offset: offset, Search: c.Query("search")}
}

// handler holds shared dependencies — like a NestJS service injected into a controller.
type handler struct {
	db     *bun.DB
	broker *stream.Broker
	cache  *flagCache
}

func newHandler(db *bun.DB, broker *stream.Broker) *handler {
	return &handler{db: db, broker: broker, cache: newFlagCache()}
}

// requireAuth is a Fiber middleware that accepts either a JWT or a project-scoped API key.
// API keys have the prefix "tfk_" and are stored as SHA-256 hashes in the api_keys table.
func (h *handler) requireAuth(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")

	if strings.HasPrefix(token, "tfk_") {
		var key db.APIKey
		err := h.db.NewSelect().Model(&key).
			Where("key_hash = ?", hashKey(token)).
			Where("expires_at IS NULL OR expires_at > ?", time.Now()).
			Scan(context.Background())
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid api key"})
		}
		c.Locals(apiKeyProjectLocal, key.ProjectID)
		// Synthesise claims so downstream role checks work.
		c.Locals("claims", &auth.Claims{Role: db.RoleOwner})
		return c.Next()
	}

	return auth.Require(c)
}

func (h *handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// checkProjectAccess returns nil if the caller is admin/superuser or is an explicit
// member of the project. Otherwise it writes a 403 and returns a non-nil error.
func (h *handler) checkProjectAccess(c *fiber.Ctx, pid int64) error {
	// API key requests are pre-scoped to a single project.
	if scopedPID, ok := c.Locals(apiKeyProjectLocal).(int64); ok {
		if scopedPID != pid {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "api key is not scoped to this project"})
		}
		return nil
	}
	claims := auth.GetClaims(c)
	if db.RoleRank(claims.Role) >= db.RoleRank(db.RoleAdmin) {
		return nil
	}
	count, err := h.db.NewSelect().TableExpr("project_members").
		Where("project_id = ? AND user_id = ?", pid, claims.UserID).
		Count(context.Background())
	if err != nil || count == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you are not a member of this project"})
	}
	return nil
}
