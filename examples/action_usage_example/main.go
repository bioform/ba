package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bioform/ba/examples/pkg/model"
	"github.com/bioform/ba/pkg/action"
	"github.com/bioform/ba/pkg/action/attr"

	"github.com/bioform/ba/examples/pkg/api"
	"github.com/bioform/ba/examples/pkg/database"
	validator "github.com/rezakhademix/govalidator/v2"
	"gorm.io/gorm"

	"github.com/bioform/ba/examples/pkg/logging"

	"github.com/bioform/ba/pkg/action/matcher/option"
	. "github.com/bioform/ba/pkg/action/matcher/tracker"
)

func main() {

	ctx := api.New(database.Default()).AddTo(context.Background())

	// This tracker should be used for test purposes only.
	action.CallTracker = NewTestTracker(action.CallTracker, NewTrackConfig(&MyAction{}, option.CallOriginal()))

	log := slog.Default()

	// Prepare the action.
	ap := action.New(ctx, &MyAction{
		SomeAttr: attr.Value(123), // Set the action-specific attribute.
	})

	// Perform the action.
	ok, err := ap.Perform()
	if !ok {

		var actionError action.ActionError
		if errors.As(err, &actionError) {
			log.Error("Action error", "error", err)
		} else {
			log.Error("Error", "error", err)
		}
	} else {
		log.Info("Action performed successfully", "SomeAttr", ap.Action().SomeAttr)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////
// Example Action implementation
/////////////////////////////////////////////////////////////////////////////////////////

// Define a specific type embedding BaseAction.
type MyAction struct {
	api.BaseAction

	SomeAttr attr.Type[int]
}

// Implement the specific behavior for MyAction.
func (a *MyAction) Perform() error {
	ctx := a.Context()
	// Put your business logic here.
	log := logging.Logger(ctx)
	log.Info("MyAction-specific perform logic")

	a.AfterCommit(func() error {
		log.Info("After commit 1")
		return errors.New("after commit 1 error")
	})

	log.Info(fmt.Sprintf("SomeAttr: %s", a.SomeAttr))

	api, err := api.From(ctx)
	if err != nil {
		return err
	}
	db := api.DB()

	user := &model.User{Name: "L1212", Email: "mmm@example.com"}
	err = db.Create(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return &model.EmailDuplicateError{Email: user.Email}
		}
		return err
	}

	var users []model.User
	db.Raw("SELECT * FROM users").Scan(&users)

	log.Info("Query users", "users", users)
	a.AfterCommit(func() error {
		log.Info("After commit 2")
		return nil
	})
	return nil
}

func (a *MyAction) IsValid() (bool, error) {
	v := validator.New()
	v.CustomRule(attr.Required(a.SomeAttr), "SomeAttr", "required")
	return v.IsPassed(), action.ErrorMap(v.Errors())
}
