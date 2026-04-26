// Variadic Options consumed by Perform / Try / IsAllowed / IsEnabled.
// SkipCache forces a fresh check; SkipTransaction skips the
// TransactionProvider wrap (used internally during nested calls);
// NopIfDisabled converts a DisabledError into (false, nil) — the behavior
// behind the Try method.
package ba

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
