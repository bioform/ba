// Package option carries the small set of configuration values the matcher
// and test tracker share: which Method to expect (Perform / Try / None),
// whether to call the original action body or stub it (CallOriginal), and
// helpers to derive the Method from a slice of ba.Options.
package option

import "github.com/bioform/ba"

type Method int

const (
	None Method = iota
	Perform
	Try
)

type TrackOptions struct {
	CallOriginal bool
	Method       Method
}

func New() TrackOptions {
	return TrackOptions{}
}

func CallOriginal() TrackOptions {
	return TrackOptions{CallOriginal: true}
}

func With(method Method) TrackOptions {
	return TrackOptions{Method: method}
}

func (opt TrackOptions) AndCallOriginal() TrackOptions {
	opt.CallOriginal = true
	return opt
}

func (opt TrackOptions) With(method Method) TrackOptions {
	opt.Method = method
	return opt
}

func (m Method) String() string {
	switch m {
	case Perform:
		return "Perform"
	case Try:
		return "Try"
	default:
		return "None"
	}
}

// Helper function to select method type
func GetMethod(opts []ba.Option) Method {
	o := ba.ParseOptions(opts)

	if o.NopIfDisabled {
		return Try
	}
	return Perform
}
