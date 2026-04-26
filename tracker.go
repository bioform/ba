// CallTracker hook plus the Tracker / Track interfaces. Every ba.New(...)
// call registers with CallTracker when it is non-nil; tests use this to
// capture, stub, or assert on action invocations without mocking the entire
// call graph. See the matcher subpackage for the test-time implementation.
package ba

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
