package action

import "context"

// CallTracker is the global tracker for all actions.
// It is used to track all actions that are performed.
// It is initialized by the package that uses the action package.
var CallTracker Tracker

type Call func() (bool, error)
type CallWrapper func() (bool, error)

type Track interface {
	PerformWrapper(ap ActionPerformer, fn Call, opts []Option) CallWrapper
}

type Tracker interface {
	Parent() Tracker
	// Track adds the action performer to the list of tracked actions.
	StartTracking(ctx context.Context, ap ActionPerformer) Track
}
