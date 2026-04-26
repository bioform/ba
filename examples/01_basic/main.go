// Minimal "hello world" of the ba library: one action, one DB write, one
// after-commit callback. Demonstrates the four core lifecycle methods and the
// transaction wrapping.
package main

import (
	"context"
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

	ok, err := ba.New(ctx, &CreateUser{
		Name:  attr.Value("Ada Lovelace"),
		Email: attr.Value("ada@example.com"),
	}).Perform()
	if !ok {
		slog.Error("CreateUser failed", "error", err)
		return
	}
	slog.Info("CreateUser succeeded")
}

type CreateUser struct {
	api.BaseAction

	Name  attr.Type[string]
	Email attr.Type[string]
}

func (a *CreateUser) IsValid() (bool, error) {
	v := validator.New()
	v.CustomRule(attr.Required(a.Name), "Name", "required")
	v.CustomRule(attr.Required(a.Email), "Email", "required")
	return v.IsPassed(), ba.ErrorMap(v.Errors())
}

func (a *CreateUser) Perform() error {
	user := &model.User{Name: a.Name.Val(), Email: a.Email.Val()}
	if err := a.DB().Create(user).Error; err != nil {
		return err
	}

	a.AfterCommit(func() error {
		slog.Info("user committed", "id", user.ID, "email", user.Email)
		return nil
	})

	return nil
}
