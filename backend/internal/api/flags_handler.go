package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"

	"toggleflow/internal/db"
)

var validFlagTypes = map[string]bool{
	"boolean": true,
	"string":  true,
	"number":  true,
	"json":    true,
}

// Variation is one possible value a flag can serve.
// Value is RawMessage so any JSON type (bool, string, number, object) round-trips
// without losing type info — similar to storing a JSON column in TypeORM with type: 'json'.
type Variation struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

type FlagEnvState struct {
	EnvironmentID    int64  `json:"environment_id"`
	EnvironmentName  string `json:"environment_name"`
	EnvironmentSlug  string `json:"environment_slug"`
	Enabled          bool   `json:"enabled"`
	DefaultVariation int    `json:"default_variation"`
}

// FlagResponse is the API shape — expands the raw DB model with parsed variations
// and per-environment state. Not embedding db.Flag avoids the field name collision
// between Flag.Variations (string) and our []Variation field.
type FlagResponse struct {
	ID           int64          `json:"id"`
	ProjectID    int64          `json:"project_id"`
	Key          string         `json:"key"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	FlagType     string         `json:"flag_type"`
	Variations   []Variation    `json:"variations"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Environments []FlagEnvState `json:"environments"`
}

func parseFlagResponse(flag db.Flag, envStates []FlagEnvState) FlagResponse {
	var variations []Variation
	if flag.Variations != "" && flag.Variations != "[]" {
		_ = json.Unmarshal([]byte(flag.Variations), &variations)
	}
	if variations == nil {
		variations = []Variation{}
	}
	return FlagResponse{
		ID:           flag.ID,
		ProjectID:    flag.ProjectID,
		Key:          flag.Key,
		Name:         flag.Name,
		Description:  flag.Description,
		FlagType:     flag.FlagType,
		Variations:   variations,
		CreatedAt:    flag.CreatedAt,
		UpdatedAt:    flag.UpdatedAt,
		Environments: envStates,
	}
}

func (h *handler) ListFlags(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	pq := parsePage(c)
	ctx := context.Background()

	fq := h.db.NewSelect().Model((*db.Flag)(nil)).Where("project_id = ?", pid).OrderExpr("created_at ASC")
	if pq.Search != "" {
		fq = fq.Where("lower(name) LIKE lower(?) OR lower(key) LIKE lower(?)", "%"+pq.Search+"%", "%"+pq.Search+"%")
	}

	total, err := fq.Count(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to count flags"})
	}

	flags := make([]db.Flag, 0)
	if err := fq.Limit(pq.Limit).Offset(pq.Offset).Scan(ctx, &flags); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch flags"})
	}

	var envs []db.Environment
	if err := h.db.NewSelect().Model(&envs).Where("project_id = ?", pid).OrderExpr("created_at ASC").Scan(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch environments"})
	}

	result := make([]FlagResponse, len(flags))

	if len(flags) > 0 {
		flagIDs := make([]int64, len(flags))
		for i, f := range flags {
			flagIDs[i] = f.ID
		}

		var flagEnvs []db.FlagEnvironment
		_ = h.db.NewSelect().Model(&flagEnvs).Where("flag_id IN (?)", bun.In(flagIDs)).Scan(ctx)

		type feKey struct{ flagID, envID int64 }
		type feState struct {
			enabled          bool
			defaultVariation int
		}
		feMap := make(map[feKey]feState)
		for _, fe := range flagEnvs {
			feMap[feKey{fe.FlagID, fe.EnvironmentID}] = feState{fe.Enabled, fe.DefaultVariation}
		}

		for i, flag := range flags {
			states := make([]FlagEnvState, len(envs))
			for j, env := range envs {
				s := feMap[feKey{flag.ID, env.ID}]
				states[j] = FlagEnvState{
					EnvironmentID:    env.ID,
					EnvironmentName:  env.Name,
					EnvironmentSlug:  env.Slug,
					Enabled:          s.enabled,
					DefaultVariation: s.defaultVariation,
				}
			}
			result[i] = parseFlagResponse(flag, states)
		}
	}

	return c.JSON(Page[FlagResponse]{Data: result, Total: total, Limit: pq.Limit, Offset: pq.Offset})
}

type createFlagRequest struct {
	Name        string      `json:"name"`
	Key         string      `json:"key"`
	Description string      `json:"description"`
	FlagType    string      `json:"flag_type"`
	Variations  []Variation `json:"variations"`
}

var defaultVariations = map[string][]Variation{
	"boolean": {
		{Name: "Enabled", Value: json.RawMessage(`true`)},
		{Name: "Disabled", Value: json.RawMessage(`false`)},
	},
	"string": {
		{Name: "Variation A", Value: json.RawMessage(`""`)},
		{Name: "Variation B", Value: json.RawMessage(`""`)},
	},
	"number": {
		{Name: "Variation A", Value: json.RawMessage(`0`)},
		{Name: "Variation B", Value: json.RawMessage(`1`)},
	},
	"json": {
		{Name: "Variation A", Value: json.RawMessage(`{}`)},
		{Name: "Variation B", Value: json.RawMessage(`{}`)},
	},
}

func (h *handler) CreateFlag(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	var req createFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if req.FlagType == "" {
		req.FlagType = "boolean"
	}
	if !validFlagTypes[req.FlagType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "flag_type must be boolean, string, number, or json"})
	}
	if len(req.Variations) == 0 {
		req.Variations = defaultVariations[req.FlagType]
	}
	if len(req.Variations) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "at least 2 variations are required"})
	}

	key := req.Key
	if key == "" {
		key = slugify(req.Name)
	}

	ctx := context.Background()

	var existing db.Flag
	if err := h.db.NewSelect().Model(&existing).Where("project_id = ? AND key = ?", pid, key).Scan(ctx); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "a flag with that key already exists"})
	}

	variationsJSON, err := json.Marshal(req.Variations)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to encode variations"})
	}

	flag := &db.Flag{
		ProjectID:   pid,
		Key:         key,
		Name:        req.Name,
		Description: req.Description,
		FlagType:    req.FlagType,
		Variations:  string(variationsJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if _, err := h.db.NewInsert().Model(flag).Exec(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create flag"})
	}

	// Seed flag_environment rows for every existing environment
	var envs []db.Environment
	if err := h.db.NewSelect().Model(&envs).Where("project_id = ?", pid).Scan(ctx); err == nil {
		for _, env := range envs {
			fe := &db.FlagEnvironment{FlagID: flag.ID, EnvironmentID: env.ID}
			_, _ = h.db.NewInsert().Model(fe).Exec(ctx)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(parseFlagResponse(*flag, []FlagEnvState{}))
}

type toggleFlagRequest struct {
	EnvironmentID    int64 `json:"environment_id"`
	Enabled          bool  `json:"enabled"`
	DefaultVariation int   `json:"default_variation"`
}

func (h *handler) UpdateFlag(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	ctx := context.Background()
	var flag db.Flag
	if err := h.db.NewSelect().Model(&flag).Where("project_id = ? AND key = ?", pid, c.Params("key")).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
	}

	var req toggleFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	var fe db.FlagEnvironment
	if err := h.db.NewSelect().Model(&fe).Where("flag_id = ? AND environment_id = ?", flag.ID, req.EnvironmentID).Scan(ctx); err != nil {
		fe = db.FlagEnvironment{FlagID: flag.ID, EnvironmentID: req.EnvironmentID, Enabled: req.Enabled, DefaultVariation: req.DefaultVariation}
		_, _ = h.db.NewInsert().Model(&fe).Exec(ctx)
	} else {
		fe.Enabled = req.Enabled
		fe.DefaultVariation = req.DefaultVariation
		_, _ = h.db.NewUpdate().Model(&fe).Column("enabled", "default_variation").Where("id = ?", fe.ID).Exec(ctx)
	}

	return c.JSON(fiber.Map{"ok": true})
}

func (h *handler) GetFlag(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	var flag db.Flag
	if err := h.db.NewSelect().Model(&flag).Where("project_id = ? AND key = ?", pid, c.Params("key")).Scan(context.Background()); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
	}
	return c.JSON(parseFlagResponse(flag, nil))
}

func (h *handler) DeleteFlag(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	ctx := context.Background()
	var flag db.Flag
	if err := h.db.NewSelect().Model(&flag).Where("project_id = ? AND key = ?", pid, c.Params("key")).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
	}

	_, _ = h.db.NewDelete().TableExpr("flag_environments").Where("flag_id = ?", flag.ID).Exec(ctx)

	if _, err := h.db.NewDelete().Model(&flag).Where("id = ?", flag.ID).Exec(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete flag"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
