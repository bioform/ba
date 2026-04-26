// Demonstrates the override hooks beyond the four core lifecycle checks:
//
//   - Init             — set defaults on action attributes
//   - IsEnabled        — subject-state preconditions (Granite analog)
//   - ErrorHandler     — translate framework/ORM errors into domain errors
//   - AfterCommit (failure) — callback errors do NOT roll back the transaction
//
// Three runs in sequence:
//   1. Happy path — Init fills a missing Name; AfterCommit deliberately
//      errors; the user row survives because AfterCommit runs post-commit.
//   2. Duplicate email — ErrorHandler translates GORM's ErrDuplicatedKey into
//      a domain-specific *EmailDuplicateError.
//   3. Signups disabled — IsEnabled rejects via ba.ErrorMap; the framework
//      surfaces it as a DisabledError.
package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bioform/ba"
	"github.com/bioform/ba/attr"
	"github.com/bioform/ba/examples/pkg/api"
	"github.com/bioform/ba/examples/pkg/database"
	"github.com/bioform/ba/examples/pkg/model"

	"gorm.io/gorm"
)

// SignupsEnabled is a stand-in for a feature flag or maintenance toggle.
var SignupsEnabled = true

func main() {
	ctx := api.New(database.Default()).AddTo(context.Background())

	run(ctx, "1. happy path (Init fills default Name; AfterCommit errors but row survives)",
		&CreateUser{Email: attr.Value("ada@example.com")})

	run(ctx, "2. duplicate email (ErrorHandler translates ErrDuplicatedKey)",
		&CreateUser{Email: attr.Value("ada@example.com")})

	SignupsEnabled = false
	run(ctx, "3. signups disabled (IsEnabled returns ErrorMap → DisabledError)",
		&CreateUser{Email: attr.Value("bob@example.com")})
}

func run(ctx context.Context, label string, action *CreateUser) {
	slog.Info(label)
	ok, err := ba.New(ctx, action).Perform()
	if !ok {
		slog.Error("  result", "error", err)
		return
	}
	slog.Info("  result: ok")
}

type CreateUser struct {
	api.BaseAction

	Name  attr.Type[string]
	Email attr.Type[string]
}

// Init runs once when ba.New constructs the performer, before any lifecycle
// check. Use it to set defaults on attributes that weren't provided.
func (a *CreateUser) Init() {
	if !a.Name.IsSet() {
		a.Name = attr.Value("Anonymous")
	}
}

// IsEnabled is the home for subject-state preconditions. Returning a
// ba.ErrorMap surfaces the rejection as a DisabledError to the caller.
func (a *CreateUser) IsEnabled() (bool, error) {
	if !SignupsEnabled {
		return false, ba.ErrorMap{"signups": "user signups are temporarily disabled"}
	}
	return true, nil
}

// ErrorHandler runs if Perform (or any lifecycle method) returns an error,
// before the framework wraps it in ActionError. Use it to translate ORM /
// infrastructure errors into typed domain errors.
func (a *CreateUser) ErrorHandler(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return &model.EmailDuplicateError{Email: a.Email.Val()}
	}
	return err
}

func (a *CreateUser) Perform() error {
	user := &model.User{Name: a.Name.Val(), Email: a.Email.Val()}
	if err := a.DB().Create(user).Error; err != nil {
		return err
	}

	a.AfterCommit(func() error {
		// AfterCommit runs *after* the transaction commits — returning an
		// error here does NOT roll back the row. The framework wraps this
		// into a CallbackError so callers know a post-commit hook failed.
		return errors.New("deliberate after-commit failure")
	})

	return nil
}
