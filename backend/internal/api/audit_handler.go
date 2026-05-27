package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"toggleflow/internal/auth"
	"toggleflow/internal/db"
)

func (h *handler) ListAudit(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	pq := parsePage(c)
	ctx := context.Background()

	q := h.db.NewSelect().Model((*db.AuditEntry)(nil)).
		Where("project_id = ?", pid).
		OrderExpr("created_at DESC")

	if resource := c.Query("resource"); resource != "" {
		q = q.Where("resource = ?", resource)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to count audit entries"})
	}

	entries := make([]db.AuditEntry, 0)
	if err := q.Limit(pq.Limit).Offset(pq.Offset).Scan(ctx, &entries); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch audit entries"})
	}

	return c.JSON(Page[db.AuditEntry]{Data: entries, Total: total, Limit: pq.Limit, Offset: pq.Offset})
}

// writeAudit records a change. Errors are silently dropped — audit failures must
// never break the primary operation.
func (h *handler) writeAudit(projectID int64, actor, action, resource, oldVal, newVal string) {
	entry := &db.AuditEntry{
		ProjectID: projectID,
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		OldValue:  oldVal,
		NewValue:  newVal,
		CreatedAt: time.Now(),
	}
	_, _ = h.db.NewInsert().Model(entry).Exec(context.Background())
}

// actorName resolves the display name for the current request's user.
func (h *handler) actorName(c *fiber.Ctx) string {
	claims := auth.GetClaims(c)
	if claims == nil {
		return "unknown"
	}
	var user db.User
	if err := h.db.NewSelect().Model(&user).Column("name").Where("id = ?", claims.UserID).Scan(context.Background()); err != nil {
		return "unknown"
	}
	return user.Name
}

// toJSON marshals v to a JSON string, returning "" on error.
func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
