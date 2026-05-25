package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
)

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

// --- Projects ---
func (h *handler) CreateProject(c *fiber.Ctx) error  { return stub(c) }
func (h *handler) ListProjects(c *fiber.Ctx) error   { return stub(c) }

// --- Environments ---
func (h *handler) CreateEnvironment(c *fiber.Ctx) error { return stub(c) }
func (h *handler) ListEnvironments(c *fiber.Ctx) error  { return stub(c) }

// --- Flags ---
func (h *handler) CreateFlag(c *fiber.Ctx) error { return stub(c) }
func (h *handler) ListFlags(c *fiber.Ctx) error  { return stub(c) }
func (h *handler) GetFlag(c *fiber.Ctx) error    { return stub(c) }
func (h *handler) UpdateFlag(c *fiber.Ctx) error { return stub(c) }
func (h *handler) DeleteFlag(c *fiber.Ctx) error { return stub(c) }

// --- Audit ---
func (h *handler) ListAudit(c *fiber.Ctx) error { return stub(c) }

// --- SDK ---
func (h *handler) SDKGetFlags(c *fiber.Ctx) error { return stub(c) }
func (h *handler) SDKEvaluate(c *fiber.Ctx) error { return stub(c) }
func (h *handler) SDKStream(c *fiber.Ctx) error   { return stub(c) }

func stub(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented"})
}
