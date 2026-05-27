package db

import (
	"time"

	"github.com/uptrace/bun"
)

// Think of these as TypeORM entities — each struct maps to a database table.
// bun:"table:x" sets the table name. bun:"pk,autoincrement" is like @PrimaryGeneratedColumn().

type Project struct {
	bun.BaseModel `bun:"table:projects"`
	ID            int64     `bun:"id,pk,autoincrement"                          json:"id"`
	Name          string    `bun:"name,notnull"                                 json:"name"`
	Key           string    `bun:"key,notnull,unique"                           json:"key"`
	Description   string    `bun:"description,notnull,default:''"               json:"description"`
	CreatedBy     *int64    `bun:"created_by"                                   json:"created_by,omitempty"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type Environment struct {
	bun.BaseModel `bun:"table:environments"`
	ID            int64     `bun:"id,pk,autoincrement"                          json:"id"`
	ProjectID     int64     `bun:"project_id,notnull"                           json:"project_id"`
	Name          string    `bun:"name,notnull"                                 json:"name"`
	Key           string    `bun:"key,notnull"                                  json:"key"`
	Description   string    `bun:"description,notnull,default:''"               json:"description"`
	Protected     bool      `bun:"protected,notnull,default:false"              json:"protected"`
	SDKKey        string    `bun:"sdk_key,notnull,unique"                       json:"-"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type Flag struct {
	bun.BaseModel `bun:"table:flags"`
	ID            int64     `bun:"id,pk,autoincrement"                          json:"id"`
	ProjectID     int64     `bun:"project_id,notnull"                           json:"project_id"`
	Key           string    `bun:"key,notnull"                                  json:"key"`
	Name          string    `bun:"name,notnull"                                 json:"name"`
	Description   string    `bun:"description"                                  json:"description"`
	FlagType      string    `bun:"flag_type,notnull,default:'boolean'"           json:"flag_type"`
	Variations    string    `bun:"variations,type:text,notnull,default:'[]'"    json:"-"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// FlagEnvironment holds per-environment state for a flag.
// Rules are stored as a JSON string — Go has no native JSON column type.
type FlagEnvironment struct {
	bun.BaseModel    `bun:"table:flag_environments"`
	ID               int64  `bun:"id,pk,autoincrement"                 json:"id"`
	FlagID           int64  `bun:"flag_id,notnull"                     json:"flag_id"`
	EnvironmentID    int64  `bun:"environment_id,notnull"              json:"environment_id"`
	Enabled          bool   `bun:"enabled,notnull,default:false"       json:"enabled"`
	Rules            string `bun:"rules,type:text"                     json:"rules"`
	DefaultVariation int    `bun:"default_variation,notnull,default:0" json:"default_variation"`
}

// Role defines what a user can do in the system.
// Ordered from highest to lowest privilege.
type Role string

const (
	RoleSuperuser Role = "superuser"
	RoleAdmin     Role = "admin"
	RoleOwner     Role = "owner"
	RoleEditor    Role = "editor"
	RoleViewer    Role = "viewer"
)

// RoleRank returns a numeric rank so we can compare roles.
// Higher number = more privilege.
func RoleRank(r Role) int {
	switch r {
	case RoleSuperuser:
		return 5
	case RoleAdmin:
		return 4
	case RoleOwner:
		return 3
	case RoleEditor:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}

type User struct {
	bun.BaseModel         `bun:"table:users"`
	ID                    int64      `bun:"id,pk,autoincrement"                          json:"id"`
	Name                  string     `bun:"name,notnull"                                 json:"name"`
	Email                 string     `bun:"email,notnull,unique"                         json:"email"`
	UUID                  string     `bun:"uuid,notnull,default:''"                      json:"uuid"`
	PasswordHash          string     `bun:"password_hash,notnull"                        json:"-"`
	Role                  Role       `bun:"role,notnull"                                 json:"role"`
	Locale                string     `bun:"locale,notnull,default:'en'"                  json:"locale"`
	WelcomeToken          string     `bun:"welcome_token,notnull"                        json:"-"`
	WelcomeTokenExpiresAt *time.Time `bun:"welcome_token_expires_at,nullzero"            json:"-"`
	ResetToken            string     `bun:"reset_token,notnull"                          json:"-"`
	ResetTokenExpiresAt   *time.Time `bun:"reset_token_expires_at,nullzero"              json:"-"`
	ActivatedAt           *time.Time `bun:"activated_at,nullzero"                        json:"activated_at,omitempty"`
	CreatedBy             *int64     `bun:"created_by"                                   json:"created_by,omitempty"`
	CreatedAt             time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt             time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type ProjectMember struct {
	bun.BaseModel `bun:"table:project_members"`
	ProjectID     int64     `bun:"project_id,pk,notnull"                        json:"project_id"`
	UserID        int64     `bun:"user_id,pk,notnull"                           json:"user_id"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type SDKKey struct {
	bun.BaseModel `bun:"table:sdk_keys"`
	ID            int64      `bun:"id,pk,autoincrement"                          json:"id"`
	EnvironmentID int64      `bun:"environment_id,notnull"                       json:"environment_id"`
	Label         string     `bun:"label,notnull"                                json:"label"`
	KeyHash       string     `bun:"key_hash,notnull,unique"                      json:"-"`
	KeyPrefix     string     `bun:"key_prefix,notnull"                           json:"key_prefix"`
	ExpiresAt     *time.Time `bun:"expires_at,nullzero"                          json:"expires_at,omitempty"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type APIKey struct {
	bun.BaseModel `bun:"table:api_keys"`
	ID            int64      `bun:"id,pk,autoincrement"                          json:"id"`
	ProjectID     int64      `bun:"project_id,notnull"                           json:"project_id"`
	Label         string     `bun:"label,notnull"                                json:"label"`
	KeyHash       string     `bun:"key_hash,notnull,unique"                      json:"-"`
	KeyPrefix     string     `bun:"key_prefix,notnull"                           json:"key_prefix"`
	ExpiresAt     *time.Time `bun:"expires_at,nullzero"                          json:"expires_at,omitempty"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type Segment struct {
	bun.BaseModel `bun:"table:segments"`
	ID            int64     `bun:"id,pk,autoincrement"                          json:"id"`
	ProjectID     int64     `bun:"project_id,notnull"                           json:"project_id"`
	Name          string    `bun:"name,notnull"                                 json:"name"`
	Key           string    `bun:"key,notnull"                                  json:"key"`
	Values        string    `bun:"values,type:text,notnull,default:'[]'"        json:"-"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type AuditEntry struct {
	bun.BaseModel `bun:"table:audit_entries"`
	ID            int64     `bun:"id,pk,autoincrement"                          json:"id"`
	ProjectID     int64     `bun:"project_id,notnull"                           json:"project_id"`
	Actor         string    `bun:"actor,notnull"                                json:"actor"`
	Action        string    `bun:"action,notnull"                               json:"action"`
	Resource      string    `bun:"resource,notnull"                             json:"resource"`
	OldValue      string    `bun:"old_value,type:text"                          json:"old_value"`
	NewValue      string    `bun:"new_value,type:text"                          json:"new_value"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
