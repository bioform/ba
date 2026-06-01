package ba_test

import (
	"errors"
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/dummy"
	. "github.com/onsi/gomega"
)

// callbackAction registers an after-commit callback during Perform and records
// whether that callback ran.
type callbackAction struct {
	ba.BaseAction

	provider    ba.TransactionProvider
	callback    func() error
	callbackRan *bool
}

func (a *callbackAction) Perform() error {
	a.AfterCommit(func() error {
		*a.callbackRan = true
		return a.callback()
	})
	return nil
}

func (a *callbackAction) TransactionProvider() ba.TransactionProvider { return a.provider }

func TestAfterCommitSkippedWhenTransactionAborts(t *testing.T) {
	g := NewWithT(t)

	ran := false
	a := &callbackAction{
		provider:    dummy.FailingTransactionProvider{Err: errors.New("commit failed")},
		callback:    func() error { return nil },
		callbackRan: &ran,
	}

	_, err := ba.New(t.Context(), a).Perform()

	// The transaction did not commit, so the callback must not have run.
	g.Expect(err).To(HaveOccurred())
	g.Expect(ran).To(BeFalse())
}

func TestAfterCommitRunsWhenTransactionCommits(t *testing.T) {
	g := NewWithT(t)

	ran := false
	a := &callbackAction{
		provider:    dummy.TransactionProvider{},
		callback:    func() error { return nil },
		callbackRan: &ran,
	}

	ok, err := ba.New(t.Context(), a).Perform()

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
	g.Expect(ran).To(BeTrue())
}

func TestAfterCommitCallbackPanicBecomesCallbackError(t *testing.T) {
	g := NewWithT(t)

	ran := false
	a := &callbackAction{
		provider:    dummy.TransactionProvider{},
		callback:    func() error { panic("callback exploded") },
		callbackRan: &ran,
	}

	ok, err := ba.New(t.Context(), a).Perform()

	g.Expect(ran).To(BeTrue())
	g.Expect(ok).To(BeFalse())
	g.Expect(err).To(HaveOccurred())

	var cbErr *ba.CallbackError
	g.Expect(errors.As(err, &cbErr)).To(BeTrue())
}
