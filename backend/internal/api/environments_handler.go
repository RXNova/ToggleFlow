package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/RXNova/ToggleFlow/internal/db"
	"github.com/gofiber/fiber/v2"
)

func (h *handler) ListEnvironments(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	envs := make([]db.Environment, 0)
	if err := h.db.NewSelect().Model(&envs).Where("project_id = ?", pid).OrderExpr("created_at ASC").Scan(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch environments"})
	}
	return c.JSON(envs)
}

type createEnvironmentRequest struct {
	Name string `json:"name"`
}

func (h *handler) CreateEnvironment(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	var req createEnvironmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	sdkKey, err := generateSDKKey()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate sdk key"})
	}

	env := &db.Environment{
		ProjectID: pid,
		Name:      req.Name,
		Slug:      slugify(req.Name),
		SDKKey:    sdkKey,
		CreatedAt: time.Now(),
	}

	if _, err := h.db.NewInsert().Model(env).Exec(context.Background()); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "an environment with that name already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to create environment"})
	}

	return c.Status(fiber.StatusCreated).JSON(env)
}

func generateSDKKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sdk_" + hex.EncodeToString(b), nil
}
