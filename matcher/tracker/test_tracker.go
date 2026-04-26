// Package tracker contains the test-only Tracker implementations used by
// the matcher package. TestTrackerImpl records action invocations in
// memory, keyed by concrete action type, so tests can assert on them after
// the triggering function returns.
//
// This file defines the TestTracker interface and the concrete
// TestTrackerImpl that records action invocations.
package tracker

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/bioform/ba"
	"github.com/bioform/ba/matcher/option"
)

type TestTracker interface {
	ba.Tracker

	Get(action ba.Action) []*TrackStats
}

type TestTrackerImpl struct {
	parent                 ba.Tracker
	trackStatsByActionType map[reflect.Type]([]*TrackStats)
}

func NewTestTracker(parent ba.Tracker, actionsToTrack ...TrackConfig) TestTracker {
	t := findTestTracker(parent, actionsToTrack)
	if t != nil {
		return NewTestTrackerProxy(parent, t, actionsToTrack...)
	}

	t = &TestTrackerImpl{
		trackStatsByActionType: make(map[reflect.Type][]*TrackStats),
		parent:                 parent,
	}
	for _, cfg := range actionsToTrack {
		stats := &TrackStats{cfg: cfg}
		t.AddTrackStats(stats)
	}
	return t
}

func (t *TestTrackerImpl) StartTracking(ctx context.Context, ap ba.ActionPerformer) ba.Track {
	action := ap.GetAction()
	actionType := typeOf(action)

	stats := t.findOrCreateStats(actionType, ap, action)

	// Handle parent tracking
	if t.parent != nil {
		parentTrack := t.parent.StartTracking(ctx, ap)
		if stats != nil {
			stats.parent = parentTrack
		} else {
			return parentTrack
		}
	}

	if stats == nil {
		return nil // Dont return a typed nil, as it will be casted to a Track interface
	}

	return stats
}

func (t *TestTrackerImpl) findOrCreateStats(actionType reflect.Type, ap ba.ActionPerformer, action ba.Action) *TrackStats {
	configs, tracked := t.isTracked(actionType)
	if !tracked {
		return nil
	}

	// Find first empty slot for tracking
	for _, cfg := range configs {
		if cfg.action == nil {
			cfg.action = action
			slog.Debug("Tracking action", "action", actionType.Name())
			return cfg
		}
	}

	// No empty slot found, mark as unexpected
	t.addUnexpectedCall(actionType, ap)
	return nil
}

func (t *TestTrackerImpl) addUnexpectedCall(actionType reflect.Type, ap ba.ActionPerformer) {
	slog.Warn("Unexpected action", "action", actionType.Name())

	stats := &TrackStats{
		cfg: TrackConfig{
			actionType: actionType,
			opt:        option.CallOriginal(),
		},
		action:     ap.GetAction(),
		unexpected: true,
	}

	t.AddTrackStats(stats)
}

func (t *TestTrackerImpl) Get(action ba.Action) []*TrackStats {
	actionType := typeOf(action)

	return t.trackStatsByActionType[actionType]
}

func (t *TestTrackerImpl) isTracked(actionType reflect.Type) ([]*TrackStats, bool) {
	stats, ok := t.trackStatsByActionType[actionType]
	return stats, ok
}

func (t *TestTrackerImpl) AddTrackStats(stats *TrackStats) {
	actionType := stats.cfg.actionType

	if _, ok := t.trackStatsByActionType[actionType]; !ok {
		t.trackStatsByActionType[actionType] = []*TrackStats{}
	}
	t.trackStatsByActionType[actionType] = append(t.trackStatsByActionType[actionType], stats)
}

func (t *TestTrackerImpl) Parent() ba.Tracker {
	return t.parent
}

func findTestTracker(parent ba.Tracker, actionsToTrack []TrackConfig) *TestTrackerImpl {
	if parent == nil {
		return nil
	}

	// Try to cast current tracker to TestTracker
	if testTracker, ok := parent.(*TestTrackerImpl); ok {
		return testTracker
	}

	return findTestTracker(parent.Parent(), actionsToTrack)
}
