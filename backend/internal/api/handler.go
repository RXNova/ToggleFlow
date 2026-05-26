package api

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"

	"toggleflow/internal/auth"
	"toggleflow/internal/db"
	"toggleflow/internal/stream"
)

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
}

func newHandler(db *bun.DB, broker *stream.Broker) *handler {
	return &handler{db: db, broker: broker}
}

func (h *handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// checkProjectAccess returns nil if the caller is admin/superuser or is an explicit
// member of the project. Otherwise it writes a 403 and returns a non-nil error.
func (h *handler) checkProjectAccess(c *fiber.Ctx, pid int64) error {
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
