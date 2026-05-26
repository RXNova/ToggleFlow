package api

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"toggleflow/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
)

func (h *handler) ListProjects(c *fiber.Ctx) error {
	pq := parsePage(c)
	ctx := context.Background()

	q := h.db.NewSelect().Model((*db.Project)(nil)).OrderExpr("created_at ASC")
	if pq.Search != "" {
		q = q.Where("lower(name) LIKE lower(?)", "%"+pq.Search+"%")
	}

	total, err := q.Count(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to count projects"})
	}

	projects := make([]db.Project, 0)
	if err := q.Limit(pq.Limit).Offset(pq.Offset).Scan(ctx, &projects); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch projects"})
	}

	return c.JSON(Page[db.Project]{Data: projects, Total: total, Limit: pq.Limit, Offset: pq.Offset})
}

type createProjectRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *handler) CreateProject(c *fiber.Ctx) error {
	var req createProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	slug := req.Slug
	if slug == "" {
		slug = slugify(req.Name)
	}

	project := &db.Project{
		Name:      req.Name,
		Slug:      slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if _, err := h.db.NewInsert().Model(project).Exec(context.Background()); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "a project with that slug already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to create project"})
	}

	return c.Status(fiber.StatusCreated).JSON(project)
}

type updateProjectRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
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
	if req.Slug != "" {
		project.Slug = req.Slug
	}
	project.UpdatedAt = time.Now()
	if _, err := h.db.NewUpdate().Model(&project).Column("name", "slug", "updated_at").Where("id = ?", pid).Exec(ctx); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "a project with that slug already exists"})
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
	h.db.NewSelect().Model(&flags).Column("id").Where("project_id = ?", pid).Scan(ctx)
	if len(flags) > 0 {
		flagIDs := make([]int64, len(flags))
		for i, f := range flags {
			flagIDs[i] = f.ID
		}
		h.db.NewDelete().TableExpr("flag_environments").Where("flag_id IN (?)", bun.In(flagIDs)).Exec(ctx)
	}
	h.db.NewDelete().TableExpr("flags").Where("project_id = ?", pid).Exec(ctx)
	h.db.NewDelete().TableExpr("environments").Where("project_id = ?", pid).Exec(ctx)
	h.db.NewDelete().TableExpr("audit_entries").Where("project_id = ?", pid).Exec(ctx)

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
