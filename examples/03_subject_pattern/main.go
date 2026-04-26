// Demonstrates the Subject pattern: an action that mutates a primary entity
// passed by reference, distinguished by convention from input parameters.
// Pairs with the Performer concept — IsAllowed reads naturally as
// "may this performer modify this subject?".
package main

import (
	"context"
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

	// Seed a user directly — not the focus of this example.
	user := &model.User{Name: "Initial", Email: "ada@example.com"}
	if err := database.Default().Create(user).Error; err != nil {
		slog.Error("seed user", "error", err)
		return
	}

	// Run the Subject-pattern action as a specific performer.
	admin := Admin{ID: 1}
	ap := ba.New(ctx, &UpdateUserName{
		User:    user,
		NewName: attr.Value("Ada Lovelace"),
	}).As(admin)

	if ok, err := ap.Perform(); !ok {
		slog.Error("UpdateUserName failed", "error", err)
		return
	}
	slog.Info("UpdateUserName succeeded", "user", user)
}

// Admin is one possible Performer type. Real apps typically use a User /
// Account / API-key struct here.
type Admin struct{ ID uint }

func (a Admin) String() string { return fmt.Sprintf("admin#%d", a.ID) }

// UpdateUserName is a Subject-pattern action: User is the entity it mutates,
// NewName is input. Distinguishing the two clarifies authorization (the
// IsAllowed check below) and makes matcher assertions express intent.
type UpdateUserName struct {
	api.BaseAction

	User    *model.User       // subject — entity being mutated
	NewName attr.Type[string] // input
}

// IsAllowed asks the question that the Subject + Performer pair makes natural:
// may this Performer modify this Subject? Real implementations would consult
// a policy / role / ownership check.
func (a *UpdateUserName) IsAllowed() (bool, error) {
	switch a.Performer().(type) {
	case Admin:
		return true, nil
	default:
		return false, nil
	}
}

func (a *UpdateUserName) IsValid() (bool, error) {
	v := validator.New()
	v.CustomRule(attr.Required(a.NewName), "NewName", "required")
	return v.IsPassed(), ba.ErrorMap(v.Errors())
}

func (a *UpdateUserName) Perform() error {
	a.User.Name = a.NewName.Val()
	return a.DB().Save(a.User).Error
}
