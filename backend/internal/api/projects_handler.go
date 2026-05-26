package api

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/RXNova/ToggleFlow/internal/db"
	"github.com/gofiber/fiber/v2"
)

func (h *handler) ListProjects(c *fiber.Ctx) error {
	projects := make([]db.Project, 0)
	err := h.db.NewSelect().Model(&projects).OrderExpr("created_at ASC").Scan(context.Background())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch projects"})
	}
	return c.JSON(projects)
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

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
