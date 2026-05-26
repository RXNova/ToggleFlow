package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"

	"toggleflow/internal/db"
	"toggleflow/internal/eval"
)

// SDKFlagConfig is the flag shape sent to SDK clients.
// It intentionally omits internal fields (project_id, created_at, etc.)
// that are irrelevant for evaluation.
type SDKFlagConfig struct {
	Key              string          `json:"key"`
	FlagType         string          `json:"flag_type"`
	Enabled          bool            `json:"enabled"`
	Variations       []Variation     `json:"variations"`
	DefaultVariation int             `json:"default_variation"`
	Rules            json.RawMessage `json:"rules"`
}

func (h *handler) SDKGetFlags(c *fiber.Ctx) error {
	sdkKey := c.Query("sdk_key")
	if sdkKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "sdk_key is required"})
	}

	ctx := context.Background()

	var env db.Environment
	if err := h.db.NewSelect().Model(&env).Where("sdk_key = ?", sdkKey).Scan(ctx); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid sdk_key"})
	}

	var flags []db.Flag
	if err := h.db.NewSelect().Model(&flags).Where("project_id = ?", env.ProjectID).OrderExpr("created_at ASC").Scan(ctx); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch flags"})
	}

	if len(flags) == 0 {
		return c.JSON([]SDKFlagConfig{})
	}

	flagIDs := make([]int64, len(flags))
	for i, f := range flags {
		flagIDs[i] = f.ID
	}

	var flagEnvs []db.FlagEnvironment
	_ = h.db.NewSelect().Model(&flagEnvs).
		Where("flag_id IN (?) AND environment_id = ?", bun.In(flagIDs), env.ID).
		Scan(ctx)

	type feKey struct{ flagID int64 }
	feMap := make(map[feKey]db.FlagEnvironment)
	for _, fe := range flagEnvs {
		feMap[feKey{fe.FlagID}] = fe
	}

	result := make([]SDKFlagConfig, len(flags))
	for i, flag := range flags {
		var variations []Variation
		if flag.Variations != "" && flag.Variations != "[]" {
			_ = json.Unmarshal([]byte(flag.Variations), &variations)
		}
		if variations == nil {
			variations = []Variation{}
		}

		fe := feMap[feKey{flag.ID}]

		rules := json.RawMessage(`[]`)
		if fe.Rules != "" {
			rules = json.RawMessage(fe.Rules)
		}

		result[i] = SDKFlagConfig{
			Key:              flag.Key,
			FlagType:         flag.FlagType,
			Enabled:          fe.Enabled,
			Variations:       variations,
			DefaultVariation: fe.DefaultVariation,
			Rules:            rules,
		}
	}

	return c.JSON(result)
}

type evaluateRequest struct {
	SDKKey  string           `json:"sdk_key"`
	FlagKey string           `json:"flag_key"`
	UserKey string           `json:"user_key"`
	Context eval.UserContext `json:"context"`
}

type evaluateResponse struct {
	FlagKey        string          `json:"flag_key"`
	Enabled        bool            `json:"enabled"`
	VariationIndex int             `json:"variation_index"`
	VariationValue json.RawMessage `json:"variation_value"`
}

func (h *handler) SDKEvaluate(c *fiber.Ctx) error {
	var req evaluateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.SDKKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "sdk_key is required"})
	}
	if req.FlagKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "flag_key is required"})
	}

	ctx := context.Background()

	var env db.Environment
	if err := h.db.NewSelect().Model(&env).Where("sdk_key = ?", req.SDKKey).Scan(ctx); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid sdk_key"})
	}

	var flag db.Flag
	if err := h.db.NewSelect().Model(&flag).Where("project_id = ? AND key = ?", env.ProjectID, req.FlagKey).Scan(ctx); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
	}

	var fe db.FlagEnvironment
	_ = h.db.NewSelect().Model(&fe).Where("flag_id = ? AND environment_id = ?", flag.ID, env.ID).Scan(ctx)

	var variations []Variation
	if flag.Variations != "" && flag.Variations != "[]" {
		_ = json.Unmarshal([]byte(flag.Variations), &variations)
	}

	variationIdx := eval.Evaluate(eval.EvalInput{
		FlagKey:          flag.Key,
		UserKey:          req.UserKey,
		UserCtx:          req.Context,
		Enabled:          fe.Enabled,
		Variations:       len(variations),
		DefaultVariation: fe.DefaultVariation,
		RulesJSON:        fe.Rules,
	})

	var value json.RawMessage
	if variationIdx >= 0 && variationIdx < len(variations) {
		value = variations[variationIdx].Value
	} else {
		value = json.RawMessage(`null`)
	}

	return c.JSON(evaluateResponse{
		FlagKey:        flag.Key,
		Enabled:        fe.Enabled,
		VariationIndex: variationIdx,
		VariationValue: value,
	})
}

func (h *handler) SDKStream(c *fiber.Ctx) error {
	sdkKey := c.Query("sdk_key")
	if sdkKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "sdk_key is required"})
	}

	ctx := context.Background()
	var env db.Environment
	if err := h.db.NewSelect().Model(&env).Where("sdk_key = ?", sdkKey).Scan(ctx); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid sdk_key"})
	}

	ch := h.broker.Subscribe()

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// SetBodyStreamWriter hands control of the response body to our function.
	// It runs in a separate goroutine and the connection stays open until we return.
	// This is the Fiber-idiomatic way to do SSE — similar to using res.write() in an
	// Express endpoint that never calls res.end().
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer h.broker.Unsubscribe(ch)

		// Send a connected heartbeat so the client knows the stream is live.
		_, _ = fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
		_ = w.Flush()

		done := c.Context().Done()

		for {
			select {
			case <-done:
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				// Only forward events for this environment's project.
				if event.ProjectID != env.ProjectID {
					continue
				}
				payload, _ := json.Marshal(map[string]any{
					"type":    "flag." + event.Action,
					"flag":    event.FlagKey,
					"env":     event.EnvKey,
					"project": event.ProjectID,
				})
				_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
				_ = w.Flush()
			}
		}
	})

	return nil
}
