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

	return nil
}
