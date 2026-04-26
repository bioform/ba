// BaseAction provides default no-op implementations of every Action method
// except Perform. Embed it in your own action structs (or in your
// application-level BaseAction) to opt into the framework with minimal
// boilerplate, overriding only the lifecycle hooks you actually need.
package ba

import (
	"context"
	"errors"
	//validator "github.com/rezakhademix/govalidator/v2"
)

type Callback func() error

// Define the base type with the shared template method.
type BaseAction struct {
	ctx       context.Context
	callbacks []Callback
	performer Performer
}

// You can ovveride this method in your action.
// It may be used to set default values for your action attributes.
func (ba *BaseAction) Init() {}

func (ba *BaseAction) SetContext(ctx context.Context) {
	ba.ctx = ctx
}

func (ba *BaseAction) Context() context.Context {
	return ba.ctx
}

// Called after the context is updated.
func (ba *BaseAction) ContextUpdated() {}

func (ba *BaseAction) Performer() Performer {
	return ba.performer
}

func (ba *BaseAction) SetPerformer(performer Performer) {
	ba.performer = performer
}

func (ba *BaseAction) IsAllowed() (bool, error) {
	return true, nil
}

func (ba *BaseAction) IsEnabled() (bool, error) {
	// See https://github.com/rezakhademix/govalidator
	// For example:
	// v := validator.New()
	// return v.IsPassed(), v.Errors()
	return true, nil
}

func (ba *BaseAction) IsValid() (bool, error) {
	// See https://github.com/rezakhademix/govalidator
	// For example:
	// v := validator.New()
	// return v.IsPassed(), v.Errors()
	return true, nil
}

func (ba *BaseAction) ErrorHandler(err error) error {
	return err
}

func (ba *BaseAction) AfterCommit(callbacks ...Callback) {
	ba.callbacks = append(ba.callbacks, callbacks...)
}

func (ba *BaseAction) AfterCommitCallback() AfterCommitCallback {
	if len(ba.callbacks) == 0 {
		return nil
	}

	callbacks := ba.callbacks // Make a copy of the callbacks.

	return func(ctx context.Context, act Action) error {
		initialContext := act.Context()

		act.SetContext(ctx)
		defer act.SetContext(initialContext)

		// The "act" embeds the BaseAction.
		var errs []error
		for _, callback := range callbacks {
			if err := callback(); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return NewCallbackError(act, errors.Join(errs...))
		}
		return nil
	}
}
