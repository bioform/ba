// Demonstrates nested business actions and the resulting nested transactions.
//
// Flow:
//   - ActionA inserts a "outer" user
//   - ActionA invokes ActionB as system (nested action → nested transaction / savepoint)
//   - ActionB inserts an "inner" user, then returns an error to force the
//     inner transaction to roll back
//   - ActionA recovers from the inner failure and returns nil, so its outer
//     transaction commits
//
// After main runs, the database contains only the "outer" user — the "inner"
// row was rolled back via the savepoint without aborting the outer
// transaction.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bioform/ba"
	"github.com/bioform/ba/attr"
	"github.com/bioform/ba/examples/pkg/api"
	"github.com/bioform/ba/examples/pkg/database"
	"github.com/bioform/ba/examples/pkg/model"

	validator "github.com/rezakhademix/govalidator/v2"
)

func main() {
	ctx := api.New(database.Default()).AddTo(context.Background())
	log := slog.Default()

	root := &ActionA{AttrA: attr.Value("Hello, World!")}
	if ok, err := ba.New(ctx, root).Perform(); !ok {
		log.Error("ActionA failed", "error", err)
		return
	}
	log.Info(fmt.Sprintf("Action %T performed successfully", root), "attrA", root.AttrA)

	var users []model.User
	database.Default().Find(&users)
	for _, u := range users {
		log.Info("surviving user", "name", u.Name, "email", u.Email)
	}
}

type ActionA struct {
	api.BaseAction

	AttrA attr.Type[string]
}

func (a *ActionA) Perform() error {
	log := slog.Default()

	if err := a.DB().Create(&model.User{Name: "outer", Email: "outer@example.com"}).Error; err != nil {
		return fmt.Errorf("outer insert: %w", err)
	}

	inner := ba.New(a.Context(), &ActionB{
		AttrB:  attr.Value(123),
		AttrB2: "some string",
	}).AsSystem()

	if ok, err := inner.Perform(); !ok {
		log.Info("inner action failed as expected — savepoint rolled back, outer continues", "error", err)
	}

	return nil
}

type ActionB struct {
	api.BaseAction

	AttrB  attr.Type[int]
	AttrB2 string
}

func (b *ActionB) Perform() error {
	if err := b.DB().Create(&model.User{Name: "inner", Email: "inner@example.com"}).Error; err != nil {
		return fmt.Errorf("inner insert: %w", err)
	}
	return errors.New("inner action chose to fail; the savepoint rolls back the inner insert")
}

func (b *ActionB) IsValid() (bool, error) {
	v := validator.New()
	v.RequiredString(b.AttrB2, "AttrB2", "required")
	return v.IsPassed(), ba.ErrorMap(v.Errors())
}
