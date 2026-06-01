package ba_test

import (
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/dummy"
	. "github.com/onsi/gomega"
)

// countingAction counts how often its IsAllowed / IsEnabled checks run, so
// tests can prove the performer memoizes them.
type countingAction struct {
	ba.BaseAction

	allowedCalls int
	enabledCalls int
}

func (a *countingAction) Perform() error                              { return nil }
func (a *countingAction) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

func (a *countingAction) IsAllowed() (bool, error) {
	a.allowedCalls++
	return true, nil
}

func (a *countingAction) IsEnabled() (bool, error) {
	a.enabledCalls++
	return true, nil
}

func TestChecksAreCached(t *testing.T) {
	cases := []struct {
		name  string
		call  func(ap *ba.ActionPerformerImpl[*countingAction], opts ...ba.Option)
		count func(a *countingAction) int
	}{
		{
			name:  "IsAllowed",
			call:  func(ap *ba.ActionPerformerImpl[*countingAction], opts ...ba.Option) { ap.IsAllowed(opts...) },
			count: func(a *countingAction) int { return a.allowedCalls },
		},
		{
			name:  "IsEnabled",
			call:  func(ap *ba.ActionPerformerImpl[*countingAction], opts ...ba.Option) { ap.IsEnabled(opts...) },
			count: func(a *countingAction) int { return a.enabledCalls },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			a := &countingAction{}
			ap := ba.New(t.Context(), a)

			// Repeated calls hit the cache: the underlying check runs once.
			tc.call(ap)
			tc.call(ap)
			tc.call(ap)
			g.Expect(tc.count(a)).To(Equal(1))

			// SkipCache bypasses the cache and forces a fresh check.
			tc.call(ap, ba.SkipCache)
			g.Expect(tc.count(a)).To(Equal(2))
		})
	}
}

func TestCacheIsPerPerformer(t *testing.T) {
	g := NewWithT(t)

	a := &countingAction{}
	ba.New(t.Context(), a).IsAllowed()

	// A second performer wrapping the same action does not share the cache.
	ba.New(t.Context(), a).IsAllowed()

	g.Expect(a.allowedCalls).To(Equal(2))
}
