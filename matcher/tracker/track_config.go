// TrackConfig — configuration for one tracked action type: the action's
// concrete reflect.Type and the matcher options (call original vs. stub,
// expected method). Constructed via NewTrackConfig.
package tracker

import (
	"reflect"

	"github.com/bioform/ba"
	"github.com/bioform/ba/matcher/option"
)

type TrackConfig struct {
	actionType reflect.Type
	opt        option.TrackOptions
}

func NewTrackConfig(action ba.Action, opt option.TrackOptions) TrackConfig {
	cfg := TrackConfig{
		actionType: typeOf(action),
		opt:        opt,
	}
	return cfg
}
