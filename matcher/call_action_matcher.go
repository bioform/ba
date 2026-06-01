// Package matcher provides Gomega matchers for asserting that a specific
// business action was invoked, optionally with specific fields, performer,
// and method (Perform vs Try). Start with CallAction(...) and chain
// AsSystem / As / With / ViaPerform / ViaTry / AndCallOriginal to refine
// the assertion. Combine matchers via NewCompositeMatcher.
//
// This file contains the primary CallActionMatcher implementation.
package matcher

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/bioform/ba"
	"github.com/bioform/ba/matcher/option"
	"github.com/bioform/ba/matcher/tracker"

	. "github.com/onsi/gomega/gstruct"
	. "github.com/onsi/gomega/types"
)

type PerformFunc[A ba.Action] func(ap *ba.ActionPerformerImpl[A]) (bool, error)

type CallActionMatcher[A ba.Action] struct {
	ExpectedAction    A
	ExpectedPerformer ba.Performer
	ExpectedParams    GomegaMatcher
	ExpectedMethod    option.Method
	CallOriginal      bool
}

// CallAction initializes the matcher with the expected action.
func CallAction[A ba.Action](expectedAction A) *CallActionMatcher[A] {
	return &CallActionMatcher[A]{
		ExpectedAction: expectedAction,
		ExpectedMethod: option.Perform,
	}
}

// AsSystem specifies that the expected performer is the system performer.
func (m *CallActionMatcher[A]) AsSystem() *CallActionMatcher[A] {
	m.ExpectedPerformer = ba.SystemPerformer
	return m
}

// As specifies the expected performer.
func (m *CallActionMatcher[A]) As(performer ba.Performer) *CallActionMatcher[A] {
	m.ExpectedPerformer = performer
	return m
}

// With specifies the expected parameters.
func (m *CallActionMatcher[A]) With(params Fields) *CallActionMatcher[A] {
	m.ExpectedParams = PointTo(MatchFields(IgnoreExtras, params))
	return m
}

// WithMethod specifies the expected method to be called.
func (m *CallActionMatcher[A]) ViaPerform() *CallActionMatcher[A] {
	m.ExpectedMethod = option.Perform
	return m
}

func (m *CallActionMatcher[A]) ViaTry() *CallActionMatcher[A] {
	m.ExpectedMethod = option.Try
	return m
}

// AndCallOriginal indicates that the original method should be called.
func (m *CallActionMatcher[A]) AndCallOriginal() *CallActionMatcher[A] {
	m.CallOriginal = true
	return m
}

func (m *CallActionMatcher[A]) Match(actual any) (bool, error) {
	var defaultAction A
	trackOpts := option.With(m.ExpectedMethod)
	if m.CallOriginal {
		trackOpts = trackOpts.AndCallOriginal()
	}

	fn, ok := actual.(func())
	if !ok {
		return false, fmt.Errorf("CallActionMatcher expects a function to execute")
	}

	trackerCfg := tracker.NewTrackConfig(defaultAction, trackOpts)
	callTracker := tracker.NewTestTracker(ba.CallTracker, trackerCfg)

	// Install our tracker for the duration of fn() only, then restore the
	// previous value. Leaving it installed would leak recorded calls into
	// later matches and other tests that share the global ba.CallTracker.
	previous := ba.CallTracker
	ba.CallTracker = callTracker
	defer func() { ba.CallTracker = previous }()

	fn()

	tracks := callTracker.Get(m.ExpectedAction)
	return m.verifyActionCalled(tracks)
}

// verifyActionCalled checks if the expected action was called with the expected parameters.
func (m *CallActionMatcher[A]) verifyActionCalled(tracks []*tracker.TrackStats) (bool, error) {
	for _, track := range tracks {
		if reflect.TypeOf(track.Action()) != reflect.TypeOf(m.ExpectedAction) {
			continue
		}
		if m.ExpectedPerformer != nil && !reflect.DeepEqual(track.Performer(), m.ExpectedPerformer) {
			slog.Warn(fmt.Sprintf("Expected performer %T, got %T", m.ExpectedPerformer, track.Performer()))
			continue
		}
		if track.Method() != m.ExpectedMethod {
			slog.Warn(fmt.Sprintf("Expected method %s, got %s", m.ExpectedMethod, track.Method()))
			continue
		}
		if m.ExpectedParams != nil {
			matched, err := m.ExpectedParams.Match(track.Action())

			if err != nil {
				return false, err
			}
			if !matched {
				continue
			}
		}
		if track.Error() != nil {
			return false, track.Error()
		}

		return true, nil
	}
	return false, nil
}

// FailureMessage returns the failure message.
func (m *CallActionMatcher[A]) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected to %s action %T as %v with parameters %v",
		m.ExpectedMethod, m.ExpectedAction, m.ExpectedPerformer, m.ExpectedParams)
}

// NegatedFailureMessage returns the negated failure message.
func (m *CallActionMatcher[A]) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Did not expect to call action %T", m.ExpectedAction)
}
