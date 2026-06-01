package ba_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/dummy"
	. "github.com/onsi/gomega"
)

type ctxKey string

const phaseKey ctxKey = "phase"

// baseTestAction embeds the library BaseAction unchanged, so its callback
// machinery operates on a real ba.Action.
type baseTestAction struct {
	ba.BaseAction
}

func (a *baseTestAction) Perform() error                              { return nil }
func (a *baseTestAction) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

func TestBaseActionDefaults(t *testing.T) {
	g := NewWithT(t)

	b := &ba.BaseAction{}

	checks := map[string]func() (bool, error){
		"IsAllowed": b.IsAllowed,
		"IsEnabled": b.IsEnabled,
		"IsValid":   b.IsValid,
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			g := NewWithT(t)
			ok, err := check()
			g.Expect(ok).To(BeTrue())
			g.Expect(err).ToNot(HaveOccurred())
		})
	}

	// ErrorHandler is a pass-through by default.
	sentinel := errors.New("boom")
	g.Expect(b.ErrorHandler(sentinel)).To(MatchError(sentinel))
}

func TestBaseActionPerformerRoundTrip(t *testing.T) {
	g := NewWithT(t)

	b := &ba.BaseAction{}
	g.Expect(b.Performer()).To(BeNil())

	b.SetPerformer("alice")
	g.Expect(b.Performer()).To(Equal("alice"))
}

func TestBaseActionContextRoundTrip(t *testing.T) {
	g := NewWithT(t)

	b := &ba.BaseAction{}
	ctx := context.WithValue(t.Context(), phaseKey, "set")
	b.SetContext(ctx)
	g.Expect(b.Context().Value(phaseKey)).To(Equal("set"))
}

func TestAfterCommitCallbackNilWhenNoCallbacks(t *testing.T) {
	g := NewWithT(t)

	a := &baseTestAction{}
	g.Expect(a.AfterCommitCallback()).To(BeNil())
}

func TestAfterCommitCallbackSwapsAndRestoresContext(t *testing.T) {
	g := NewWithT(t)

	a := &baseTestAction{}
	a.SetContext(context.WithValue(t.Context(), phaseKey, "initial"))

	var seen any
	a.AfterCommit(func() error {
		seen = a.Context().Value(phaseKey)
		return nil
	})

	cb := a.AfterCommitCallback()
	g.Expect(cb).ToNot(BeNil())

	txCtx := context.WithValue(t.Context(), phaseKey, "commit")
	g.Expect(cb(txCtx, a)).ToNot(HaveOccurred())

	// The callback runs with the transaction context...
	g.Expect(seen).To(Equal("commit"))
	// ...and the action's context is restored afterwards.
	g.Expect(a.Context().Value(phaseKey)).To(Equal("initial"))
}

func TestAfterCommitCallbackJoinsErrors(t *testing.T) {
	g := NewWithT(t)

	a := &baseTestAction{}
	a.SetContext(t.Context())
	a.AfterCommit(
		func() error { return errors.New("first failure") },
		func() error { return nil },
		func() error { return errors.New("second failure") },
	)

	err := a.AfterCommitCallback()(t.Context(), a)
	g.Expect(err).To(HaveOccurred())

	var cbErr *ba.CallbackError
	g.Expect(errors.As(err, &cbErr)).To(BeTrue())
	g.Expect(err.Error()).To(ContainSubstring("first failure"))
	g.Expect(err.Error()).To(ContainSubstring("second failure"))
}
