package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"toggleflow/internal/db"
)

type createKeyRequest struct {
	Label     string     `json:"label"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type createdKeyResponse[T any] struct {
	Key    string `json:"key"`
	Record T      `json:"record"`
}

func generateRawKey(prefix string) (raw, hash, keyPrefix string, err error) {
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = prefix + hex.EncodeToString(b)
	h := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(h[:])
	keyPrefix = raw
	if len(keyPrefix) > 12 {
		keyPrefix = keyPrefix[:12]
	}
	return
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func (h *handler) ListSDKKeys(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}
	eid, err := strconv.ParseInt(c.Params("eid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid environment id"})
	}

	keys := make([]db.SDKKey, 0)
	if err := h.db.NewSelect().Model(&keys).
		Where("environment_id = ?", eid).
		OrderExpr("created_at ASC").
		Scan(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch sdk keys"})
	}
	return c.JSON(keys)
}

func (h *handler) CreateSDKKey(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}
	eid, err := strconv.ParseInt(c.Params("eid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid environment id"})
	}

	var req createKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Label == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "label is required"})
	}

	raw, hash, prefix, err := generateRawKey("sdk_")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate key"})
	}

	record := &db.SDKKey{
		EnvironmentID: eid,
		Label:         req.Label,
		KeyHash:       hash,
		KeyPrefix:     prefix,
		ExpiresAt:     req.ExpiresAt,
		CreatedAt:     time.Now(),
	}

	if _, err := h.db.NewInsert().Model(record).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create sdk key"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdKeyResponse[db.SDKKey]{Key: raw, Record: *record})
}

func (h *handler) DeleteSDKKey(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}
	kid, err := strconv.ParseInt(c.Params("kid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid key id"})
	}

	if _, err := h.db.NewDelete().TableExpr("sdk_keys").Where("id = ?", kid).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete sdk key"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handler) ListAPIKeys(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	keys := make([]db.APIKey, 0)
	if err := h.db.NewSelect().Model(&keys).
		Where("project_id = ?", pid).
		OrderExpr("created_at ASC").
		Scan(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch api keys"})
	}
	return c.JSON(keys)
}

func (h *handler) CreateAPIKey(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	var req createKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Label == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "label is required"})
	}

	raw, hash, prefix, err := generateRawKey("tfk_")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate key"})
	}

	record := &db.APIKey{
		ProjectID: pid,
		Label:     req.Label,
		KeyHash:   hash,
		KeyPrefix: prefix,
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	}

	if _, err := h.db.NewInsert().Model(record).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create api key"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdKeyResponse[db.APIKey]{Key: raw, Record: *record})
}

func (h *handler) DeleteAPIKey(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}
	kid, err := strconv.ParseInt(c.Params("kid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid key id"})
	}

	if _, err := h.db.NewDelete().TableExpr("api_keys").Where("id = ?", kid).Exec(context.Background()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete api key"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
