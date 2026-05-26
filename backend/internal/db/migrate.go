package db

import (
	"context"

	"github.com/uptrace/bun"
)

// Migrate creates all tables if they don't exist.
// Bun's CreateTableIfNotExists is like running a TypeORM synchronize on startup.
func Migrate(db *bun.DB) error {
	ctx := context.Background()

	models := []interface{}{
		(*User)(nil),
		(*Project)(nil),
		(*Environment)(nil),
		(*Flag)(nil),
		(*FlagEnvironment)(nil),
		(*AuditEntry)(nil),
	}

	for _, model := range models {
		if _, err := db.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(ctx); err != nil {
			return err
		}
	}

	// Add columns introduced after initial schema — SQLite errors if the column already
	// exists, so we ignore those errors (equivalent to ADD COLUMN IF NOT EXISTS).
	for _, stmt := range []string{
		`ALTER TABLE flags ADD COLUMN flag_type TEXT NOT NULL DEFAULT 'boolean'`,
		`ALTER TABLE flags ADD COLUMN variations TEXT NOT NULL DEFAULT '[]'`,
		`ALTER TABLE projects ADD COLUMN created_by INTEGER REFERENCES users(id)`,
		`ALTER TABLE projects RENAME COLUMN slug TO key`,
		`ALTER TABLE projects ADD COLUMN description TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE environments ADD COLUMN description TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE environments ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`,
	} {
		_, _ = db.ExecContext(ctx, stmt)
	}

	return nil
}
