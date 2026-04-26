package api

import (
	"context"
	"fmt"

	"github.com/bioform/ba"
	"gorm.io/gorm"
)

type BaseAction struct {
	ba.BaseAction
	api API
}

func (b *BaseAction) SetContext(ctx context.Context) {
	b.BaseAction.SetContext(ctx)

	api, err := From(ctx)
	if err != nil {
		panic(fmt.Errorf("set api: %w", err))
	}

	b.api = api
}

func (b *BaseAction) TransactionProvider() ba.TransactionProvider {
	return b.api
}

func (b BaseAction) API() API {
	return b.api
}

func (b BaseAction) DB() *gorm.DB {
	return b.api.DB()
}
