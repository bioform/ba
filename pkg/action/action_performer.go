package action

import (
	"context"
	"errors"
	"slices"
)

type ActionPerformer interface {
	GetAction() Action
	Performer() Performer
	IsValid() (bool, error)
	IsPerformable(...Option) (bool, error)
}

type ActionPerformerImpl[A Action] struct {
	action A
	track  Track

	callbacks   []AfterCommitCallback
	addCallback AddCallbackFunc

	isAllowedCache *Cache
	isEnabledCache *Cache
}

func New[A Action](ctx context.Context, action A) *ActionPerformerImpl[A] {
	ap := &ActionPerformerImpl[A]{action: action}
	// Collect the created action performer somewhere to track its execution late in tests.
	if CallTracker != nil {
		ap.track = CallTracker.StartTracking(ctx, ap)
	}

	ap.setContext(ap.setAddCallback(ctx))

	action.Init()

	return ap
}

func (ap *ActionPerformerImpl[A]) Equal(other *ActionPerformerImpl[A]) bool {
	return ap == other
}

func (ap *ActionPerformerImpl[A]) As(performer Performer) *ActionPerformerImpl[A] {
	ap.action.SetPerformer(performer)
	return ap
}

func (ap *ActionPerformerImpl[A]) AsSystem() *ActionPerformerImpl[A] {
	ap.action.SetPerformer(SystemPerformer)
	return ap
}

func (ap *ActionPerformerImpl[A]) Action() A {
	return ap.action
}

func (ap *ActionPerformerImpl[A]) GetAction() Action {
	return ap.action
}

func (ap *ActionPerformerImpl[A]) Performer() Performer {
	return ap.action.Performer()
}

// Perform executes the action within a transaction context.
func (ap *ActionPerformerImpl[A]) Perform() (bool, error) {
	return ap.performWithTracking()
}

// Try executes the action but does not return an error if the action is disabled.
func (ap *ActionPerformerImpl[A]) Try() (bool, error) {
	return ap.performWithTracking(NopIfDisabled)
}

// perform executes the action within a transaction context provided by the
// TransactionProvider. It first checks if the action is enabled and valid
// before performing it. If the action is disabled and NopIfDisabled option is provided,
// it will proceed without error.
//
// Parameters:
//   - ctx: The context for the operation, used for managing request-scoped
//     values (DB reference), cancellation, and deadlines.
//   - NopIfDisabled: A boolean flag indicating whether the action should proceed
//     even if it is disabled.
//
// Returns:
//   - ok: A boolean indicating whether the action was successfully performed.
//   - err: An error if any occurred during the transaction or action execution.
func (ap *ActionPerformerImpl[A]) perform(opts []Option) (bool, error) {
	fn := func() (bool, error) {
		if ok, err := ap.IsPerformable(append(opts, SkipTransaction)...); !ok || err != nil {
			return ok, err
		}
		if ok, err := ap.IsValid(); !ok || err != nil {
			return ok, err
		}
		if err := ap.action.Perform(); err != nil {
			return false, err
		}
		return true, nil
	}
	ok, err := ap.transaction(fn)

	if ok && err == nil {
		if errs := ap.AfterCommit(); len(errs) > 0 {
			ok = false

			actionErrors := make([]error, 0, len(errs))
			for _, err := range errs {
				actionErrors = append(actionErrors, ap.wrapError(err))
			}
			err = errors.Join(actionErrors...)
		}
	}

	err = ap.handleError(err)

	return ok, err
}

// IsAllowed checks if the action is allowed and caches the result.
func (ap *ActionPerformerImpl[A]) IsAllowed(opts ...Option) (bool, error) {
	o := ParseOptions(opts)

	if !o.SkipCache && ap.isAllowedCache != nil {
		return ap.isAllowedCache.ok, ap.isAllowedCache.err
	}
	fn := func() (bool, error) {
		ok, err := ap.checkAllowed()
		ap.isAllowedCache = &Cache{ok: ok, err: err}
		return ok, err
	}

	if o.SkipTransaction {
		return fn()
	}
	return ap.transaction(fn)
}

// IsEnabled checks if the action is enabled and caches the result.
func (ap *ActionPerformerImpl[A]) IsEnabled(opts ...Option) (bool, error) {
	o := ParseOptions(opts)

	if !o.SkipCache && ap.isEnabledCache != nil {
		return ap.isEnabledCache.ok, ap.isEnabledCache.err
	}
	fn := func() (bool, error) {
		ok, err := ap.checkEnabled(opts...)
		ap.isEnabledCache = &Cache{ok: ok, err: err}
		return ok, err
	}
	if o.SkipTransaction {
		return fn()
	}
	return ap.transaction(fn)
}

// IsPerformable checks if the action is both allowed and enabled.
func (ap *ActionPerformerImpl[A]) IsPerformable(opts ...Option) (bool, error) {
	if ok, err := ap.IsAllowed(opts...); !ok || err != nil {
		return false, err
	}
	if ok, err := ap.IsEnabled(opts...); !ok || err != nil {
		return false, err
	}
	return true, nil
}

func (ap *ActionPerformerImpl[A]) checkAllowed() (bool, error) {
	ok, err := ap.action.IsAllowed()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, NewAuthorizationError(ap.action)
	}
	return true, nil
}

func (ap *ActionPerformerImpl[A]) checkEnabled(opts ...Option) (bool, error) {
	ok, err := ap.action.IsEnabled()
	if !ok {
		if errMap, ok := err.(ErrorMap); ok {
			if slices.Contains(opts, NopIfDisabled) {
				return false, nil
			}
			return false, NewDisabledError(ap.action, errMap)
		}
		return false, err
	}
	return true, nil
}

func (ap *ActionPerformerImpl[A]) IsValid() (bool, error) {
	ok, err := ap.action.IsValid()
	if !ok {
		if errMap, ok := err.(ErrorMap); ok {
			return false, NewValidationError(ap.action, errMap)
		}
		return false, err
	}
	return true, nil
}

func (ap *ActionPerformerImpl[A]) transaction(fn func() (bool, error)) (ok bool, err error) {
	tp := ap.action.TransactionProvider()
	ctx := ap.action.Context()

	err = tp.Transaction(ctx, func(txCtx context.Context) error {
		ap.setContext(txCtx)
		defer ap.setContext(ctx)

		ok, err = fn()
		return err
	})
	return ok, err
}

func (ap *ActionPerformerImpl[A]) setContext(ctx context.Context) {
	ap.action.SetContext(ctx)
	ap.action.ContextUpdated()
}

func (ap *ActionPerformerImpl[A]) handleError(err error) error {
	if err != nil {
		handledError := ap.action.ErrorHandler(err)
		if handledError != err {
			return handledError
		}

		err = ap.wrapError(err)
	}
	return err
}

func (ap *ActionPerformerImpl[A]) wrapError(err error) error {
	var actionErr ActionError
	if ok := errors.As(err, &actionErr); ok {
		var action Action = ap.action
		if actionErr.action != action {
			return NewActionError(ap.action, err)
		}
		return err
	}
	return NewActionError(ap.action, err)
}

/////////////////////////////////////////////////////////////////////////////////////////
// Tracking
/////////////////////////////////////////////////////////////////////////////////////////

func (ap *ActionPerformerImpl[A]) performWithTracking(opts ...Option) (bool, error) {
	if ap.track != nil {
		return ap.track.PerformWrapper(ap, func() (bool, error) {
			return ap.perform(opts)
		}, opts)()
	}
	return ap.perform(opts)
}
