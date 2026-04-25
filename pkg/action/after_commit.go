package action

import (
	"context"
	"errors"
	"runtime/debug"
)

type AddCallbackFunc func(callback AfterCommitCallback)
type addCallback string

var addCallbackKey addCallback = "addCallbackFunc"

func (ap *ActionPerformerImpl[A]) AfterCommit() []error {
	var (
		errs []error
		act  Action          = ap.Action()
		ctx  context.Context = act.Context()
	)

	if callback := act.AfterCommitCallback(); callback != nil {
		ap.addCallback(callback)
	}

	for _, callback := range ap.callbacks {
		if err := executeCallback(ctx, callback, act); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func executeCallback(ctx context.Context, callback AfterCommitCallback, act Action) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewCallbackErrorWithStack(act, errors.New("panic in callback"), debug.Stack())
		}
	}()
	return callback(ctx, act)
}

func (ap *ActionPerformerImpl[A]) setAddCallback(ctx context.Context) context.Context {
	if fn, ok := ctx.Value(addCallbackKey).(AddCallbackFunc); ok {
		ap.addCallback = fn
	} else {
		ap.addCallback = func(callback AfterCommitCallback) {
			ap.callbacks = append(ap.callbacks, callback)
		}
		ctx = context.WithValue(ctx, addCallbackKey, ap.addCallback)
	}
	return ctx
}
