package api

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"

	"toggleflow/internal/auth"
	"toggleflow/internal/db"
)

// ProjectResponse is the API shape for a project, including the creator's name
// resolved via a LEFT JOIN on users in ListProjects.
type ProjectResponse struct {
	ID            int64     `bun:"id"              json:"id"`
	Name          string    `bun:"name"            json:"name"`
	Key           string    `bun:"key"             json:"key"`
	Description   string    `bun:"description"     json:"description"`
	CreatedBy     *int64    `bun:"created_by"      json:"created_by,omitempty"`
	CreatedByName string    `bun:"created_by_name" json:"created_by_name"`
	CreatedAt     time.Time `bun:"created_at"      json:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"      json:"updated_at"`
}

func (h *handler) ListProjects(c *fiber.Ctx) error {
	pq := parsePage(c)
	ctx := context.Background()

	q := h.db.NewSelect().
		TableExpr("projects AS p").
		ColumnExpr("p.*").
		ColumnExpr("COALESCE(u.name, '') AS created_by_name").
		Join("LEFT JOIN users AS u ON u.id = p.created_by").
		OrderExpr("p.created_at ASC")

	if pq.Search != "" {
		q = q.Where("lower(p.name) LIKE lower(?)", "%"+pq.Search+"%")
	}

	total, err := q.Count(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to count projects"})
	}

	projects := make([]ProjectResponse, 0)
	if err := q.Limit(pq.Limit).Offset(pq.Offset).Scan(ctx, &projects); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch projects"})
	}

	return c.JSON(Page[ProjectResponse]{Data: projects, Total: total, Limit: pq.Limit, Offset: pq.Offset})
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

func (h *handler) CreateProject(c *fiber.Ctx) error {
	var req createProjectRequest
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

	claims := auth.GetClaims(c)
	project := &db.Project{
		Name:        req.Name,
		Key:         key,
		Description: req.Description,
		CreatedBy:   &claims.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if _, err := h.db.NewInsert().Model(project).Exec(context.Background()); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "a project with that key already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to create project"})
	}

	return c.Status(fiber.StatusCreated).JSON(project)
}

type updateProjectRequest struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

func (h *handler) UpdateProject(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	var req updateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	ctx := context.Background()
	var project db.Project
	if err := h.db.NewSelect().Model(&project).Where("id = ?", pid).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	project.Name = req.Name
	project.Description = req.Description
	if req.Key != "" {
		project.Key = req.Key
	}
	project.UpdatedAt = time.Now()
	if _, err := h.db.NewUpdate().Model(&project).Column("name", "key", "description", "updated_at").Where("id = ?", pid).Exec(ctx); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "a project with that key already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to update project"})
	}

	return c.JSON(project)
}

func (h *handler) DeleteProject(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	ctx := context.Background()

	// Cascade: delete flags and their environment states, then environments, then audit entries
	var flags []db.Flag
	_ = h.db.NewSelect().Model(&flags).Column("id").Where("project_id = ?", pid).Scan(ctx)
	if len(flags) > 0 {
		flagIDs := make([]int64, len(flags))
		for i, f := range flags {
			flagIDs[i] = f.ID
		}
		_, _ = h.db.NewDelete().TableExpr("flag_environments").Where("flag_id IN (?)", bun.In(flagIDs)).Exec(ctx)
	}
	_, _ = h.db.NewDelete().TableExpr("flags").Where("project_id = ?", pid).Exec(ctx)
	_, _ = h.db.NewDelete().TableExpr("environments").Where("project_id = ?", pid).Exec(ctx)
	_, _ = h.db.NewDelete().TableExpr("audit_entries").Where("project_id = ?", pid).Exec(ctx)

	if _, err := h.db.NewDelete().TableExpr("projects").Where("id = ?", pid).Exec(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete project"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
