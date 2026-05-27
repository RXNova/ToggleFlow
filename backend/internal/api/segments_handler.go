package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"toggleflow/internal/db"
)

type SegmentResponse struct {
	ID        int64           `json:"id"`
	ProjectID int64           `json:"project_id"`
	Name      string          `json:"name"`
	Key       string          `json:"key"`
	Values    json.RawMessage `json:"values"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func parseSegmentResponse(s db.Segment) SegmentResponse {
	values := json.RawMessage(`[]`)
	if s.Values != "" {
		values = json.RawMessage(s.Values)
	}
	return SegmentResponse{
		ID:        s.ID,
		ProjectID: s.ProjectID,
		Name:      s.Name,
		Key:       s.Key,
		Values:    values,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func (h *handler) ListSegments(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	ctx := context.Background()
	segments := make([]db.Segment, 0)
	if err := h.db.NewSelect().Model(&segments).Where("project_id = ?", pid).OrderExpr("name ASC").Scan(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch segments"})
	}

	result := make([]SegmentResponse, len(segments))
	for i, s := range segments {
		result[i] = parseSegmentResponse(s)
	}
	return c.JSON(result)
}

type createSegmentRequest struct {
	Name   string          `json:"name"`
	Key    string          `json:"key"`
	Values json.RawMessage `json:"values"`
}

func (h *handler) CreateSegment(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	var req createSegmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	key := req.Key
	if key == "" {
		key = slugify(req.Name)
	}

	valuesJSON := "[]"
	if len(req.Values) > 0 {
		valuesJSON = string(req.Values)
	}

	ctx := context.Background()
	var existing db.Segment
	if err := h.db.NewSelect().Model(&existing).Where("project_id = ? AND key = ?", pid, key).Scan(ctx); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "a segment with that key already exists"})
	}

	now := time.Now()
	seg := db.Segment{
		ProjectID: pid,
		Name:      req.Name,
		Key:       key,
		Values:    valuesJSON,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := h.db.NewInsert().Model(&seg).Exec(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create segment"})
	}
	return c.Status(fiber.StatusCreated).JSON(parseSegmentResponse(seg))
}

type updateSegmentRequest struct {
	Name   string          `json:"name"`
	Values json.RawMessage `json:"values"`
}

func (h *handler) UpdateSegment(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	sid, err := strconv.ParseInt(c.Params("sid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid segment id"})
	}

	var req updateSegmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	ctx := context.Background()
	var seg db.Segment
	if err := h.db.NewSelect().Model(&seg).Where("id = ? AND project_id = ?", sid, pid).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "segment not found"})
	}

	seg.Name = req.Name
	if len(req.Values) > 0 {
		seg.Values = string(req.Values)
	}
	seg.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(&seg).Column("name", "values", "updated_at").Where("id = ?", seg.ID).Exec(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update segment"})
	}
	return c.JSON(parseSegmentResponse(seg))
}

func (h *handler) DeleteSegment(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	sid, err := strconv.ParseInt(c.Params("sid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid segment id"})
	}

	ctx := context.Background()
	res, err := h.db.NewDelete().Model((*db.Segment)(nil)).Where("id = ? AND project_id = ?", sid, pid).Exec(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete segment"})
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "segment not found"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
