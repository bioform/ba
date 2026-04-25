package tracker

import (
	"github.com/bioform/ba/pkg/action"
	"github.com/bioform/ba/pkg/action/matcher/option"
)

type TrackStats struct {
	error

	cfg        TrackConfig
	parent     action.Track
	method     option.Method
	action     action.Action
	performer  action.Performer
	unexpected bool
}

func (stats *TrackStats) PerformWrapper(ap action.ActionPerformer, call action.Call, opts []action.Option) action.CallWrapper {
	stats.method = option.GetMethod(opts)
	stats.performer = ap.Performer()

	if !stats.cfg.opt.CallOriginal {
		return func() (bool, error) {
			ok, err := ap.IsPerformable(opts...)
			if !ok || err != nil {
				stats.error = err
				return ok, err
			}
			ok, err = ap.IsValid()
			stats.error = err
			return ok, err
		}
	}

	return func() (bool, error) {
		wrappedCall := func() (bool, error) {
			ok, err := call()
			stats.error = err
			return ok, err
		}

		if stats.parent != nil {
			return stats.parent.PerformWrapper(ap, wrappedCall, opts)()
		}

		return wrappedCall()
	}
}

func (stats *TrackStats) Action() action.Action {
	return stats.action
}

func (stats *TrackStats) Method() option.Method {
	return stats.method
}

func (stats *TrackStats) Performer() action.Performer {
	return stats.performer
}

func (stats *TrackStats) Error() error {
	return stats.error
}
