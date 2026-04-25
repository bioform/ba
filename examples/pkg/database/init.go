package database

import (
	"log/slog"
	"fmt"
)

var (
	Dsn string // default DB DSN
)

func init() {
	Dsn = ":memory:" // Default to in-memory SQLite database for testing

	db, err := initSqliteDB(Dsn)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		slog.Panicf("cannot open db(%s): %v", Dsn, err)
	}
	defaultDB = db
}

func initSqliteDB(dsn string) (*gorm.DB, error) {
	if strings.Contains(dsn, ":memory:") || strings.Contains(dsn, "mode=memory") {
		slog.Info("Restore schema(in-memory DB)", slog.String("dsn", dsn))
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	err = schema.Restore(db)
	if err != nil {
		return nil, fmt.Errorf("failed to restore schema: %w", err)
	}
	return db, nil
}
