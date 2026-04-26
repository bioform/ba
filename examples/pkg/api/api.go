package api

import (
	"context"

	"github.com/bioform/ba"
	"gorm.io/gorm"
)

type API interface {
	ba.TransactionProvider
	DB() *gorm.DB
}

type api struct {
	db *gorm.DB
}

func New(db *gorm.DB) *api {
	return &api{db: db}
}

func (a *api) DB() *gorm.DB {
	return a.db
}

// This part is required for the ba.TransactionProvider interface.
// It allows the API to provide a transaction context for actions.
func (a *api) Transaction(ctx context.Context, lambda func(newContext context.Context) error) error {
	return a.db.Transaction(func(tx *gorm.DB) error {
		newAPI := *a
		newAPI.db = tx
		return lambda(newAPI.AddTo(ctx))
	})
}
