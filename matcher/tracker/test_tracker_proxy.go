// TestTrackerProxy lets multiple TestTracker setups coexist by chaining
// through a parent TestTrackerImpl. Used when a test installs additional
// expectations on top of an existing tracker without losing the previous
// ones.
package tracker

import (
	"context"
	"log/slog"

	"github.com/bioform/ba"
)

type TestTrackerProxy struct {
	parent ba.Tracker

	testTracker *TestTrackerImpl
}

func NewTestTrackerProxy(parent ba.Tracker, testTracker *TestTrackerImpl, actionsToTrack ...TrackConfig) *TestTrackerProxy {
	if testTracker == nil {
		slog.Error("TestTrackerProxy: testTracker is nil")
		return nil
	}

	t := &TestTrackerProxy{
		parent:      parent,
		testTracker: testTracker,
	}

	for _, cfg := range actionsToTrack {
		stats := &TrackStats{cfg: cfg}
		testTracker.AddTrackStats(stats)
	}

	return t
}

func (p *TestTrackerProxy) Get(action ba.Action) []*TrackStats {
	return p.testTracker.Get(action)
}

func (p *TestTrackerProxy) Parent() ba.Tracker {
	return p.parent
}

func (p *TestTrackerProxy) StartTracking(ctx context.Context, ap ba.ActionPerformer) ba.Track {
	return p.testTracker.StartTracking(ctx, ap)
}
