package database

import (
	"fmt"
	"log/slog"

	"github.com/bioform/ba/examples/pkg/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Dsn string

func init() {
	Dsn = ":memory:"

	db, err := initSqliteDB(Dsn)
	if err != nil {
		slog.Error("cannot open db", "dsn", Dsn, "error", err)
		panic(err)
	}
	defaultDB = db
}

func initSqliteDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}
	return db, nil
}
