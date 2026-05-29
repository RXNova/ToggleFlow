package api

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

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

func (h *handler) sdkEnvironment(c *fiber.Ctx, sdkKey string) (*db.Environment, error) {
	if sdkKey == "" {
		_ = c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "sdk_key is required"})
		return nil, fiber.ErrUnauthorized
	}

	keyHash := hashKey(sdkKey)

	if env := h.keyCache.get(keyHash); env != nil {
		return env, nil
	}

	ctx := context.Background()
	var key db.SDKKey
	if err := h.db.NewSelect().Model(&key).
		Where("key_hash = ?", keyHash).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Scan(ctx); err != nil {
		_ = c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid sdk_key"})
		return nil, fiber.ErrUnauthorized
	}
	var env db.Environment
	if err := h.db.NewSelect().Model(&env).Where("id = ?", key.EnvironmentID).Scan(ctx); err != nil {
		_ = c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid sdk_key"})
		return nil, fiber.ErrUnauthorized
	}

	h.keyCache.set(keyHash, &env, key.ExpiresAt)
	return &env, nil
}

func (h *handler) SDKGetFlags(c *fiber.Ctx) error {
	env, err := h.sdkEnvironment(c, c.Query("sdk_key"))
	if err != nil {
		return nil
	}

	data := h.cache.get(env.ProjectID, env.ID)
	if data == nil {
		ctx := context.Background()
		var flags []db.Flag
		if err := h.db.NewSelect().Model(&flags).Where("project_id = ?", env.ProjectID).OrderExpr("created_at ASC").Scan(ctx); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to fetch flags"})
		}

		var result []SDKFlagConfig
		if len(flags) > 0 {
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

			result = make([]SDKFlagConfig, len(flags))
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
		} else {
			result = []SDKFlagConfig{}
		}

		var merr error
		data, merr = json.Marshal(result)
		if merr != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to encode flags"})
		}
		h.cache.set(env.ProjectID, env.ID, data)
	}

	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(data))
	if c.Get("If-None-Match") == etag {
		return c.SendStatus(fiber.StatusNotModified)
	}
	c.Set("ETag", etag)
	c.Set("Content-Type", "application/json")
	return c.Send(data)
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

	env, err := h.sdkEnvironment(c, req.SDKKey)
	if err != nil {
		return nil
	}

	ctx := context.Background()

	var variations []Variation
	var enabled bool
	var defaultVariation int
	var rulesJSON string
	found := false

	if cached := h.cache.get(env.ProjectID, env.ID); cached != nil {
		var configs []SDKFlagConfig
		if json.Unmarshal(cached, &configs) == nil {
			for _, cfg := range configs {
				if cfg.Key == req.FlagKey {
					variations = cfg.Variations
					enabled = cfg.Enabled
					defaultVariation = cfg.DefaultVariation
					if cfg.Rules != nil {
						rulesJSON = string(cfg.Rules)
					}
					found = true
					break
				}
			}
			if !found {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
			}
		}
	}

	if !found {
		var flag db.Flag
		if err := h.db.NewSelect().Model(&flag).Where("project_id = ? AND key = ?", env.ProjectID, req.FlagKey).Scan(ctx); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "flag not found"})
		}
		var fe db.FlagEnvironment
		_ = h.db.NewSelect().Model(&fe).Where("flag_id = ? AND environment_id = ?", flag.ID, env.ID).Scan(ctx)
		if flag.Variations != "" && flag.Variations != "[]" {
			_ = json.Unmarshal([]byte(flag.Variations), &variations)
		}
		enabled = fe.Enabled
		defaultVariation = fe.DefaultVariation
		rulesJSON = fe.Rules
	}

	var dbSegments []db.Segment
	_ = h.db.NewSelect().Model(&dbSegments).Where("project_id = ?", env.ProjectID).Scan(ctx)
	segments := make(map[string][]any, len(dbSegments))
	for _, s := range dbSegments {
		var vals []any
		if s.Values != "" {
			_ = json.Unmarshal([]byte(s.Values), &vals)
		}
		segments[s.Key] = vals
	}

	variationIdx := eval.Evaluate(eval.EvalInput{
		FlagKey:          req.FlagKey,
		UserKey:          req.UserKey,
		UserCtx:          req.Context,
		Enabled:          enabled,
		Variations:       len(variations),
		DefaultVariation: defaultVariation,
		RulesJSON:        rulesJSON,
		Segments:         segments,
	})

	var value json.RawMessage
	if variationIdx >= 0 && variationIdx < len(variations) {
		value = variations[variationIdx].Value
	} else {
		value = json.RawMessage(`null`)
	}

	return c.JSON(evaluateResponse{
		FlagKey:        req.FlagKey,
		Enabled:        enabled,
		VariationIndex: variationIdx,
		VariationValue: value,
	})
}

func (h *handler) SDKStream(c *fiber.Ctx) error {
	env, err := h.sdkEnvironment(c, c.Query("sdk_key"))
	if err != nil {
		return nil
	}

	ch := h.broker.Subscribe(env.ProjectID)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// SetBodyStreamWriter hands control of the response body to our function.
	// It runs in a separate goroutine and the connection stays open until we return.
	// This is the Fiber-idiomatic way to do SSE — similar to using res.write() in an
	// Express endpoint that never calls res.end().
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer h.broker.Unsubscribe(env.ProjectID, ch)

		_, _ = fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
		_ = w.Flush()

		done := c.Context().Done()
		heartbeat := time.NewTicker(30 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-done:
				return
			case <-heartbeat.C:
				_, _ = fmt.Fprintf(w, ": ping\n\n")
				_ = w.Flush()
			case event, ok := <-ch:
				if !ok {
					return
				}
				var payload []byte
				if event.Action == "demo.reset" {
					payload, _ = json.Marshal(map[string]any{"type": "demo.reset"})
				} else {
					payload, _ = json.Marshal(map[string]any{
						"type":    "flag." + event.Action,
						"flag":    event.FlagKey,
						"env":     event.EnvKey,
						"project": event.ProjectID,
					})
				}
				_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
				_ = w.Flush()
			}
		}
	})

	return nil
}
