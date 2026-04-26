// TrackStats — recorded outcome of a tracked action call: the action
// instance, performer, method invoked, any error, and whether the call was
// expected. Implements ba.Track via PerformWrapper, which wires the
// stub-vs-original choice into the action's execution.
package tracker

import (
	"github.com/bioform/ba"
	"github.com/bioform/ba/matcher/option"
)

type TrackStats struct {
	error

	cfg        TrackConfig
	parent     ba.Track
	method     option.Method
	action     ba.Action
	performer  ba.Performer
	unexpected bool
}

func (stats *TrackStats) PerformWrapper(ap ba.ActionPerformer, call ba.Call, opts []ba.Option) ba.CallWrapper {
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

func (stats *TrackStats) Action() ba.Action {
	return stats.action
}

func (stats *TrackStats) Method() option.Method {
	return stats.method
}

func (stats *TrackStats) Performer() ba.Performer {
	return stats.performer
}

func (stats *TrackStats) Error() error {
	return stats.error
}
