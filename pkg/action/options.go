package action

type Option int

const (
	SkipCache Option = iota
	SkipTransaction
	NopIfDisabled
)

type Options struct {
	SkipCache       bool
	SkipTransaction bool
	NopIfDisabled   bool
}

func ParseOptions(opts []Option) Options {
	if len(opts) == 0 {
		return Options{}
	}

	var o Options
	for _, opt := range opts {
		switch opt {
		case SkipCache:
			o.SkipCache = true
		case SkipTransaction:
			o.SkipTransaction = true
		case NopIfDisabled:
			o.NopIfDisabled = true
		}
	}
	return o
}
