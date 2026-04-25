package tracker

import (
	"reflect"

	"github.com/bioform/ba/pkg/action"
	"github.com/bioform/ba/pkg/action/matcher/option"
)

type TrackConfig struct {
	actionType reflect.Type
	opt        option.TrackOptions
}

func NewTrackConfig(action action.Action, opt option.TrackOptions) TrackConfig {
	cfg := TrackConfig{
		actionType: typeOf(action),
		opt:        opt,
	}
	return cfg
}
