package db

import (
	"time"

	"github.com/uptrace/bun"
)

// Think of these as TypeORM entities — each struct maps to a database table.
// bun:"table:x" sets the table name. bun:"pk,autoincrement" is like @PrimaryGeneratedColumn().

type Project struct {
	bun.BaseModel `bun:"table:projects"`
	ID          int64     `bun:"id,pk,autoincrement"`
	Name        string    `bun:"name,notnull"`
	Slug        string    `bun:"slug,notnull,unique"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

type Environment struct {
	bun.BaseModel `bun:"table:environments"`
	ID        int64     `bun:"id,pk,autoincrement"`
	ProjectID int64     `bun:"project_id,notnull"`
	Name      string    `bun:"name,notnull"`
	Slug      string    `bun:"slug,notnull"`
	SDKKey    string    `bun:"sdk_key,notnull,unique"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

type Flag struct {
	bun.BaseModel `bun:"table:flags"`
	ID          int64     `bun:"id,pk,autoincrement"`
	ProjectID   int64     `bun:"project_id,notnull"`
	Key         string    `bun:"key,notnull"`
	Name        string    `bun:"name,notnull"`
	Description string    `bun:"description"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

// FlagEnvironment holds per-environment state for a flag — enabled, variations, targeting rules.
// Variations and Rules are stored as JSON strings (Go has no native JSON column type).
type FlagEnvironment struct {
	bun.BaseModel `bun:"table:flag_environments"`
	ID               int64  `bun:"id,pk,autoincrement"`
	FlagID           int64  `bun:"flag_id,notnull"`
	EnvironmentID    int64  `bun:"environment_id,notnull"`
	Enabled          bool   `bun:"enabled,notnull,default:false"`
	Variations       string `bun:"variations,type:text"`
	Rules            string `bun:"rules,type:text"`
	DefaultVariation int    `bun:"default_variation,notnull,default:0"`
}

type AuditEntry struct {
	bun.BaseModel `bun:"table:audit_entries"`
	ID        int64     `bun:"id,pk,autoincrement"`
	ProjectID int64     `bun:"project_id,notnull"`
	Actor     string    `bun:"actor,notnull"`
	Action    string    `bun:"action,notnull"`
	Resource  string    `bun:"resource,notnull"`
	OldValue  string    `bun:"old_value,type:text"`
	NewValue  string    `bun:"new_value,type:text"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
