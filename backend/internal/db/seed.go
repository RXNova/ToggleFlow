package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"
)

// Reset wipes all data so Seed can repopulate from scratch.
func Reset(database *bun.DB) error {
	ctx := context.Background()
	tables := []string{
		"audit_entries", "flag_environments", "flags", "segments",
		"sdk_keys", "api_keys", "project_members", "environments", "projects", "users",
	}
	for _, t := range tables {
		if _, err := database.ExecContext(ctx, "DELETE FROM "+t); err != nil {
			return err
		}
	}
	return nil
}

// Seed inserts demo data if the database is empty. Idempotent — safe to call on every startup.
func Seed(database *bun.DB) error {
	ctx := context.Background()

	count, _ := database.NewSelect().Model((*User)(nil)).Count(ctx)
	if count > 0 {
		return nil
	}

	now := time.Now()
	activated := now

	hash, err := bcrypt.GenerateFromPassword([]byte("flag-demo-flow"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// --- Users ---
	demo := &User{
		UUID:         uuid.NewString(),
		Name:         "Demo User",
		Email:        "demo@toggleflow.io",
		PasswordHash: string(hash),
		Role:         RoleSuperuser,
		ActivatedAt:  &activated,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := database.NewInsert().Model(demo).Exec(ctx); err != nil {
		return err
	}

	alice := &User{
		UUID:         uuid.NewString(),
		Name:         "Alice Chen",
		Email:        "alice@acme.com",
		PasswordHash: string(hash),
		Role:         RoleEditor,
		ActivatedAt:  &activated,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, _ = database.NewInsert().Model(alice).Exec(ctx)

	bob := &User{
		UUID:         uuid.NewString(),
		Name:         "Bob Kumar",
		Email:        "bob@acme.com",
		PasswordHash: string(hash),
		Role:         RoleViewer,
		ActivatedAt:  &activated,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, _ = database.NewInsert().Model(bob).Exec(ctx)

	// --- Project ---
	project := &Project{
		Name:        "Acme App",
		Key:         "acme-app",
		Description: "Main web application for Acme Corp",
		CreatedBy:   &demo.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := database.NewInsert().Model(project).Exec(ctx); err != nil {
		return err
	}

	for _, uid := range []int64{demo.ID, alice.ID, bob.ID} {
		m := &ProjectMember{ProjectID: project.ID, UserID: uid, CreatedAt: now}
		_, _ = database.NewInsert().Model(m).Exec(ctx)
	}

	// --- Environments ---
	type envRow struct {
		name, key, desc string
		protected       bool
	}
	envDefs := []envRow{
		{"Production", "production", "Live user traffic — protected", true},
		{"Staging", "staging", "Pre-release integration testing", false},
		{"Development", "development", "Local and CI development", false},
	}

	envs := make([]*Environment, 0, len(envDefs))
	for _, d := range envDefs {
		raw, keyHash, keyPrefix := seedKey("sdk_")
		env := &Environment{
			ProjectID:   project.ID,
			Name:        d.name,
			Key:         d.key,
			Description: d.desc,
			Protected:   d.protected,
			SDKKey:      raw,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if _, err := database.NewInsert().Model(env).Exec(ctx); err != nil {
			return err
		}
		envs = append(envs, env)
		_, _ = database.NewInsert().Model(&SDKKey{
			EnvironmentID: env.ID,
			Label:         "Default",
			KeyHash:       keyHash,
			KeyPrefix:     keyPrefix,
			CreatedAt:     now,
		}).Exec(ctx)
	}

	prod, staging, dev := envs[0], envs[1], envs[2]

	// --- API key ---
	_, apiHash, apiPrefix := seedKey("tfk_")
	_, _ = database.NewInsert().Model(&APIKey{
		ProjectID: project.ID,
		Label:     "CI / CD",
		KeyHash:   apiHash,
		KeyPrefix: apiPrefix,
		CreatedAt: now,
	}).Exec(ctx)

	// --- Segments ---
	betaSeg := &Segment{
		ProjectID: project.ID,
		Name:      "Beta Users",
		Key:       "beta-users",
		Values:    j([]string{"alice@acme.com", "bob@acme.com", "carol@acme.com", "dave@acme.com"}),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, _ = database.NewInsert().Model(betaSeg).Exec(ctx)

	proSeg := &Segment{
		ProjectID: project.ID,
		Name:      "Pro Plan",
		Key:       "pro-plan",
		Values:    j([]string{"pro", "enterprise"}),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, _ = database.NewInsert().Model(proSeg).Exec(ctx)

	countrySeg := &Segment{
		ProjectID: project.ID,
		Name:      "EU Countries",
		Key:       "eu-countries",
		Values:    j([]string{"DE", "FR", "NL", "ES", "IT", "PL", "SE"}),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, _ = database.NewInsert().Model(countrySeg).Exec(ctx)

	// --- Flags ---

	// 1. boolean — dark-mode
	darkMode := insertFlag(ctx, database, project.ID, "dark-mode", "Dark Mode",
		"Enable dark mode UI", "boolean",
		j([]map[string]any{{"name": "On", "value": true}, {"name": "Off", "value": false}}), now)

	seedFE(ctx, database, darkMode, prod.ID, false, 1, "")
	seedFE(ctx, database, darkMode, staging.ID, true, 0, "")
	seedFE(ctx, database, darkMode, dev.ID, true, 0, "")

	// 2. boolean — new-dashboard (with beta segment rule)
	betaRule := j([]map[string]any{{
		"conditions": []map[string]any{{
			"attribute": "email",
			"operator":  "in",
			"segment":   "beta-users",
		}},
		"serve": 0,
	}})
	newDash := insertFlag(ctx, database, project.ID, "new-dashboard", "New Dashboard",
		"Redesigned analytics dashboard", "boolean",
		j([]map[string]any{{"name": "New", "value": true}, {"name": "Legacy", "value": false}}), now)

	seedFE(ctx, database, newDash, prod.ID, true, 1, betaRule)
	seedFE(ctx, database, newDash, staging.ID, true, 0, "")
	seedFE(ctx, database, newDash, dev.ID, true, 0, "")

	// 3. boolean — pricing-experiment (percentage rollout)
	rolloutRule := j([]map[string]any{{
		"conditions": []map[string]any{},
		"rollout": []map[string]any{
			{"variation": 0, "weight": 20},
			{"variation": 1, "weight": 80},
		},
	}})
	pricing := insertFlag(ctx, database, project.ID, "pricing-experiment", "Pricing Experiment",
		"A/B test new pricing page — 20% see new pricing", "boolean",
		j([]map[string]any{{"name": "New Pricing", "value": true}, {"name": "Current Pricing", "value": false}}), now)

	seedFE(ctx, database, pricing, prod.ID, true, 1, rolloutRule)
	seedFE(ctx, database, pricing, staging.ID, true, 1, rolloutRule)
	seedFE(ctx, database, pricing, dev.ID, false, 1, "")

	// 4. string — welcome-message
	welcomeMsg := insertFlag(ctx, database, project.ID, "welcome-message", "Welcome Message",
		"Personalised banner message shown on login", "string",
		j([]map[string]any{
			{"name": "Default", "value": "Welcome back!"},
			{"name": "Pro", "value": "Welcome back, Pro member!"},
			{"name": "Beta", "value": "Thanks for being a beta tester!"},
		}), now)

	proRule := j([]map[string]any{{
		"conditions": []map[string]any{{
			"attribute": "plan",
			"operator":  "in",
			"segment":   "pro-plan",
		}},
		"serve": 1,
	}})
	seedFE(ctx, database, welcomeMsg, prod.ID, true, 0, proRule)
	seedFE(ctx, database, welcomeMsg, staging.ID, true, 0, "")
	seedFE(ctx, database, welcomeMsg, dev.ID, true, 0, "")

	// 5. number — max-upload-mb
	maxUpload := insertFlag(ctx, database, project.ID, "max-upload-mb", "Max Upload Size (MB)",
		"Maximum file upload size in megabytes", "number",
		j([]map[string]any{
			{"name": "10 MB", "value": 10},
			{"name": "100 MB", "value": 100},
			{"name": "500 MB", "value": 500},
		}), now)

	seedFE(ctx, database, maxUpload, prod.ID, true, 0, "")
	seedFE(ctx, database, maxUpload, staging.ID, true, 1, "")
	seedFE(ctx, database, maxUpload, dev.ID, true, 2, "")

	// 6. json — checkout-config
	checkoutCfg := insertFlag(ctx, database, project.ID, "checkout-config", "Checkout Config",
		"Feature toggles and limits for the checkout flow", "json",
		j([]map[string]any{
			{"name": "Standard", "value": map[string]any{"show_promo": false, "max_items": 10, "express_enabled": false}},
			{"name": "Enhanced", "value": map[string]any{"show_promo": true, "max_items": 50, "express_enabled": true}},
		}), now)

	seedFE(ctx, database, checkoutCfg, prod.ID, true, 0, "")
	seedFE(ctx, database, checkoutCfg, staging.ID, true, 1, "")
	seedFE(ctx, database, checkoutCfg, dev.ID, true, 1, "")

	// --- Audit entries ---
	audits := []*AuditEntry{
		{ProjectID: project.ID, Actor: "Demo User", Action: "flag.created", Resource: "dark-mode", OldValue: "", NewValue: `{"name":"Dark Mode","key":"dark-mode","type":"boolean"}`, CreatedAt: now.Add(-48 * time.Hour)},
		{ProjectID: project.ID, Actor: "Demo User", Action: "flag.toggled", Resource: "dark-mode", OldValue: `{"env":"Production","enabled":false}`, NewValue: `{"env":"Production","enabled":false}`, CreatedAt: now.Add(-47 * time.Hour)},
		{ProjectID: project.ID, Actor: "Alice Chen", Action: "flag.created", Resource: "new-dashboard", OldValue: "", NewValue: `{"name":"New Dashboard","key":"new-dashboard","type":"boolean"}`, CreatedAt: now.Add(-24 * time.Hour)},
		{ProjectID: project.ID, Actor: "Alice Chen", Action: "flag.rules_updated", Resource: "new-dashboard", OldValue: `{"env":"Production","rules":"[]"}`, NewValue: `{"env":"Production","rules":"[beta rule]"}`, CreatedAt: now.Add(-23 * time.Hour)},
		{ProjectID: project.ID, Actor: "Demo User", Action: "flag.created", Resource: "pricing-experiment", OldValue: "", NewValue: `{"name":"Pricing Experiment","key":"pricing-experiment","type":"boolean"}`, CreatedAt: now.Add(-12 * time.Hour)},
		{ProjectID: project.ID, Actor: "Demo User", Action: "flag.toggled", Resource: "pricing-experiment", OldValue: `{"env":"Production","enabled":false}`, NewValue: `{"env":"Production","enabled":true}`, CreatedAt: now.Add(-11 * time.Hour)},
		{ProjectID: project.ID, Actor: "Bob Kumar", Action: "flag.toggled", Resource: "dark-mode", OldValue: `{"env":"Staging","enabled":false}`, NewValue: `{"env":"Staging","enabled":true}`, CreatedAt: now.Add(-2 * time.Hour)},
	}
	for _, a := range audits {
		_, _ = database.NewInsert().Model(a).Exec(ctx)
	}

	return nil
}

func insertFlag(ctx context.Context, database *bun.DB, pid int64, key, name, desc, flagType, variations string, now time.Time) *Flag {
	f := &Flag{
		ProjectID:   pid,
		Key:         key,
		Name:        name,
		Description: desc,
		FlagType:    flagType,
		Variations:  variations,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, _ = database.NewInsert().Model(f).Exec(ctx)
	return f
}

func seedFE(ctx context.Context, database *bun.DB, flag *Flag, envID int64, enabled bool, defaultVar int, rules string) {
	if rules == "" {
		rules = "[]"
	}
	_, _ = database.NewInsert().Model(&FlagEnvironment{
		FlagID:           flag.ID,
		EnvironmentID:    envID,
		Enabled:          enabled,
		DefaultVariation: defaultVar,
		Rules:            rules,
	}).Exec(ctx)
}

func seedKey(prefix string) (raw, keyHash, keyPrefix string) {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	raw = prefix + hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	keyHash = hex.EncodeToString(sum[:])
	keyPrefix = raw
	if len(keyPrefix) > 12 {
		keyPrefix = keyPrefix[:12]
	}
	return
}

func j(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
