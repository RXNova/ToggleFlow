package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
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
	db *bun.DB
}

func newHandler(db *bun.DB) *handler {
	return &handler{db: db}
}

func (h *handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// --- Audit ---
func (h *handler) ListAudit(c *fiber.Ctx) error { return stub(c) }

// --- SDK ---
func (h *handler) SDKGetFlags(c *fiber.Ctx) error { return stub(c) }
func (h *handler) SDKEvaluate(c *fiber.Ctx) error { return stub(c) }
func (h *handler) SDKStream(c *fiber.Ctx) error   { return stub(c) }

func stub(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented"})
}
