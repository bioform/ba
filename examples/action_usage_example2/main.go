package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bioform/ba/pkg/action"
	"github.com/bioform/ba/pkg/action/attr"
	"github.com/bioform/ba/examples/pkg/api"
	"github.com/bioform/ba/examples/pkg/database"

	validator "github.com/rezakhademix/govalidator/v2"
)

func main() {
	ctx := api.New(database.Default(), email.Client()).AddTo(context.Background())

	log := logging.Logger(ctx)

	// Perform the action.
	myAction := &ActionA{AttrA: attr.Value("Hello, World!")}
	ok, err := action.New(ctx, myAction).Perform()
	if !ok {
		log.Error("Error", "error", err)
	} else {
		log.Info(fmt.Sprintf("Action %T performed successfully", myAction), "attrA", myAction.AttrA)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////
// Example Action implementation
/////////////////////////////////////////////////////////////////////////////////////////

// Define a specific type embedding BaseAction.
type ActionA struct {
	api.BaseAction

	AttrA attr.Type[string]
}

// Implement the specific behavior for ActionA.
// ActionA performs ActionB as a system action.
func (a *ActionA) Perform() error {
	ctx := a.Context()
	log := slog.Default()

	actionPerformer := action.New(ctx, &ActionB{
		AttrB:  attr.Value(123), // Set the action-specific attribute.
		AttrB2: "some string",
	})
	actionPerformer = actionPerformer.AsSystem()
	ok, err := actionPerformer.Perform()

	if !ok {
		log.Error("Error", "error", err)
	} else {
		log.Info(fmt.Sprintf("Action %T performed successfully", actionPerformer.Action()), "attrB", actionPerformer.Action().AttrB)
	}

	return nil
}

type ActionB struct {
	api.BaseAction

	AttrB  attr.Type[int]
	AttrB2 string
}

// Implement the specific behavior for ActionB.
func (b *ActionB) Perform() error {

	return nil
}

func (b *ActionB) IsValid() (bool, error) {
	v := validator.New()
	v.RequiredString(b.AttrB2, "AttrB2", "required")
	return v.IsPassed(), action.ErrorMap(v.Errors())
}
