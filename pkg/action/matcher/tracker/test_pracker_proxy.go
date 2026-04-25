package tracker

import (
	"context"
	"log/slog"

	"github.com/bioform/ba/pkg/action"
)

type TestTrackerProxy struct {
	parent action.Tracker

	testTracker *TestTrackerImpl
}

func NewTestTrackerProxy(parent action.Tracker, testTracker *TestTrackerImpl, actionsToTrack ...TrackConfig) *TestTrackerProxy {
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

func (p *TestTrackerProxy) Get(action action.Action) []*TrackStats {
	return p.testTracker.Get(action)
}

func (p *TestTrackerProxy) Parent() action.Tracker {
	return p.parent
}

func (p *TestTrackerProxy) StartTracking(ctx context.Context, ap action.ActionPerformer) action.Track {
	return p.testTracker.StartTracking(ctx, ap)
}
