package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"

	"toggleflow/internal/auth"
	"toggleflow/internal/db"
	"toggleflow/internal/stream"
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
	EnvironmentID    int64           `json:"environment_id"`
	EnvironmentName  string          `json:"environment_name"`
	EnvironmentKey   string          `json:"environment_key"`
	Protected        bool            `json:"protected"`
	Enabled          bool            `json:"enabled"`
	DefaultVariation int             `json:"default_variation"`
	Rules            json.RawMessage `json:"rules"`
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
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
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
			rules            string
		}
		feMap := make(map[feKey]feState)
		for _, fe := range flagEnvs {
			feMap[feKey{fe.FlagID, fe.EnvironmentID}] = feState{fe.Enabled, fe.DefaultVariation, fe.Rules}
		}

		for i, flag := range flags {
			states := make([]FlagEnvState, len(envs))
			for j, env := range envs {
				s := feMap[feKey{flag.ID, env.ID}]
				rules := json.RawMessage(`[]`)
				if s.rules != "" {
					rules = json.RawMessage(s.rules)
				}
				states[j] = FlagEnvState{
					EnvironmentID:    env.ID,
					EnvironmentName:  env.Name,
					EnvironmentKey:   env.Key,
					Protected:        env.Protected,
					Enabled:          s.enabled,
					DefaultVariation: s.defaultVariation,
					Rules:            rules,
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
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
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

	h.writeAudit(pid, h.actorName(c), "flag.created", flag.Key, "",
		toJSON(map[string]any{"name": flag.Name, "key": flag.Key, "type": flag.FlagType}))

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

type updateFlagRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Variations  []Variation `json:"variations"`
}

func (h *handler) UpdateFlag(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	ctx := context.Background()
	var flag db.Flag
	if err := h.db.NewSelect().Model(&flag).Where("project_id = ? AND key = ?", pid, c.Params("key")).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
	}

	var req updateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if len(req.Variations) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "at least 2 variations are required"})
	}

	variationsJSON, err := json.Marshal(req.Variations)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to encode variations"})
	}

	oldName, oldDesc := flag.Name, flag.Description
	flag.Name = req.Name
	flag.Description = req.Description
	flag.Variations = string(variationsJSON)
	flag.UpdatedAt = time.Now()

	if _, err := h.db.NewUpdate().Model(&flag).Column("name", "description", "variations", "updated_at").Where("id = ?", flag.ID).Exec(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update flag"})
	}

	h.writeAudit(pid, h.actorName(c), "flag.updated", flag.Key,
		toJSON(map[string]any{"name": oldName, "description": oldDesc}),
		toJSON(map[string]any{"name": flag.Name, "description": flag.Description}))

	h.broker.Publish(stream.Event{ProjectID: pid, FlagKey: flag.Key, Action: "updated"})
	h.cache.bustProject(pid)

	return c.JSON(parseFlagResponse(flag, nil))
}

type toggleFlagRequest struct {
	EnvironmentID    int64 `json:"environment_id"`
	Enabled          bool  `json:"enabled"`
	DefaultVariation int   `json:"default_variation"`
}

func (h *handler) ToggleFlagEnv(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
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

	var env db.Environment
	if err := h.db.NewSelect().Model(&env).Where("id = ?", req.EnvironmentID).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "environment not found"})
	}
	if env.Protected {
		claims := auth.GetClaims(c)
		if db.RoleRank(claims.Role) < db.RoleRank(db.RoleAdmin) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "this environment is protected — only admins can change flag states"})
		}
	}

	var fe db.FlagEnvironment
	oldEnabled := false
	if err := h.db.NewSelect().Model(&fe).Where("flag_id = ? AND environment_id = ?", flag.ID, req.EnvironmentID).Scan(ctx); err != nil {
		fe = db.FlagEnvironment{FlagID: flag.ID, EnvironmentID: req.EnvironmentID, Enabled: req.Enabled, DefaultVariation: req.DefaultVariation}
		_, _ = h.db.NewInsert().Model(&fe).Exec(ctx)
	} else {
		oldEnabled = fe.Enabled
		fe.Enabled = req.Enabled
		fe.DefaultVariation = req.DefaultVariation
		_, _ = h.db.NewUpdate().Model(&fe).Column("enabled", "default_variation").Where("id = ?", fe.ID).Exec(ctx)
	}

	h.writeAudit(pid, h.actorName(c), "flag.toggled", flag.Key,
		toJSON(map[string]any{"env": env.Name, "enabled": oldEnabled}),
		toJSON(map[string]any{"env": env.Name, "enabled": req.Enabled}))

	h.broker.Publish(stream.Event{
		ProjectID: pid,
		EnvKey:    env.Key,
		FlagKey:   flag.Key,
		Action:    "updated",
	})
	h.cache.bust(pid, req.EnvironmentID)

	return c.JSON(fiber.Map{"ok": true})
}

func (h *handler) GetFlag(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
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
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
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

	h.writeAudit(pid, h.actorName(c), "flag.deleted", flag.Key,
		toJSON(map[string]any{"name": flag.Name, "key": flag.Key}), "")

	h.broker.Publish(stream.Event{
		ProjectID: pid,
		FlagKey:   flag.Key,
		Action:    "deleted",
	})
	h.cache.bustProject(pid)

	return c.SendStatus(fiber.StatusNoContent)
}

type saveRulesRequest struct {
	EnvironmentID int64           `json:"environment_id"`
	Rules         json.RawMessage `json:"rules"`
}

func (h *handler) SaveFlagRules(c *fiber.Ctx) error {
	pid, err := strconv.ParseInt(c.Params("pid"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}
	if err := h.checkProjectAccess(c, pid); err != nil {
		return err
	}

	ctx := context.Background()
	var flag db.Flag
	if err := h.db.NewSelect().Model(&flag).Where("project_id = ? AND key = ?", pid, c.Params("key")).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
	}

	var req saveRulesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	rulesJSON := string(req.Rules)
	if rulesJSON == "" {
		rulesJSON = "[]"
	}

	var env db.Environment
	if err := h.db.NewSelect().Model(&env).Where("id = ?", req.EnvironmentID).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "environment not found"})
	}

	var fe db.FlagEnvironment
	if err := h.db.NewSelect().Model(&fe).Where("flag_id = ? AND environment_id = ?", flag.ID, req.EnvironmentID).Scan(ctx); err != nil {
		fe = db.FlagEnvironment{FlagID: flag.ID, EnvironmentID: req.EnvironmentID, Rules: rulesJSON}
		_, _ = h.db.NewInsert().Model(&fe).Exec(ctx)
	} else {
		oldRules := fe.Rules
		fe.Rules = rulesJSON
		_, _ = h.db.NewUpdate().Model(&fe).Column("rules").Where("id = ?", fe.ID).Exec(ctx)

		h.writeAudit(pid, h.actorName(c), "flag.rules_updated", flag.Key,
			toJSON(map[string]any{"env": env.Name, "rules": oldRules}),
			toJSON(map[string]any{"env": env.Name, "rules": rulesJSON}))
	}

	h.broker.Publish(stream.Event{ProjectID: pid, EnvKey: env.Key, FlagKey: flag.Key, Action: "updated"})
	h.cache.bust(pid, req.EnvironmentID)

	return c.JSON(fiber.Map{"ok": true})
}
