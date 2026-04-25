package option

import "github.com/bioform/ba/pkg/action"

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
func GetMethod(opts []action.Option) Method {
	o := action.ParseOptions(opts)

	if o.NopIfDisabled {
		return Try
	}
	return Perform
}
