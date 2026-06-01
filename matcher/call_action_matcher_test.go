package matcher_test

import (
	"errors"
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/attr"
	"github.com/bioform/ba/dummy"
	. "github.com/bioform/ba/matcher"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

// childAction is a minimal real action used to exercise the matcher. It records
// whether its body ran (performed) so tests can tell stubbed calls (body
// skipped) apart from AndCallOriginal calls (body run), and returns err.
type childAction struct {
	ba.BaseAction

	Name      string
	Num       attr.Type[int]
	err       error
	performed bool
}

func (a *childAction) Perform() error {
	a.performed = true
	return a.err
}
func (a *childAction) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

type otherAction struct {
	ba.BaseAction
}

func (a *otherAction) Perform() error                              { return nil }
func (a *otherAction) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

// resetTracker isolates each test from the global ba.CallTracker the matcher
// installs (and, prior to the save/restore fix, leaves behind). Because
// ba.CallTracker is process-global mutable state, tests using it must NOT call
// t.Parallel().
func resetTracker(t *testing.T) {
	t.Helper()
	ba.CallTracker = nil
	t.Cleanup(func() { ba.CallTracker = nil })
}

func TestMatchStubsBodyByDefault(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	m := CallAction(&childAction{}).
		AsSystem().
		With(Fields{
			"Name": Equal("hi"),
			"Num":  Equal(attr.Value(7)),
		}).
		ViaPerform()

	// err is non-nil, but because the body is stubbed it never runs, so the
	// match still succeeds.
	a := &childAction{
		Name: "hi",
		Num:  attr.Value(7),
		err:  errors.New("body must not run when stubbed"),
	}
	ok, err := m.Match(func() {
		ba.New(t.Context(), a).AsSystem().Perform()
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
	g.Expect(a.performed).To(BeFalse(), "stubbed call must not run the action body")
}

func TestMatchAndCallOriginalRunsBody(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	m := CallAction(&childAction{}).AsSystem().AndCallOriginal()

	a := &childAction{Name: "real"}
	ok, err := m.Match(func() {
		ba.New(t.Context(), a).AsSystem().Perform()
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
	g.Expect(a.performed).To(BeTrue(), "AndCallOriginal must run the real action body")
}

func TestMatchAndCallOriginalRejectsFailingBody(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	// AndCallOriginal runs the real body and asserts it succeeded; a failing
	// body must surface as a matcher error.
	m := CallAction(&childAction{}).AsSystem().AndCallOriginal()

	ok, err := m.Match(func() {
		ba.New(t.Context(), &childAction{err: errors.New("kaboom")}).AsSystem().Perform()
	})

	g.Expect(err).To(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestMatchMethod(t *testing.T) {
	cases := []struct {
		name      string
		expectTry bool
		callTry   bool
		want      bool
	}{
		{"expect Perform, called via Perform", false, false, true},
		{"expect Try, called via Try", true, true, true},
		{"expect Perform, called via Try", false, true, false},
		{"expect Try, called via Perform", true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			resetTracker(t)

			m := CallAction(&childAction{})
			if tc.expectTry {
				m = m.ViaTry()
			} else {
				m = m.ViaPerform()
			}

			ok, err := m.Match(func() {
				ap := ba.New(t.Context(), &childAction{})
				if tc.callTry {
					ap.Try()
				} else {
					ap.Perform()
				}
			})

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ok).To(Equal(tc.want))
		})
	}
}

func TestMatchPerformer(t *testing.T) {
	type user struct{ id int }
	alice := &user{1}
	bob := &user{2}

	cases := []struct {
		name     string
		expect   ba.Performer
		actual   ba.Performer
		asSystem bool
		want     bool
	}{
		{"matching performer", alice, alice, false, true},
		{"different performer", alice, bob, false, false},
		{"system vs user", alice, nil, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			resetTracker(t)

			m := CallAction(&childAction{}).As(tc.expect)

			ok, err := m.Match(func() {
				ap := ba.New(t.Context(), &childAction{})
				if tc.asSystem {
					ap.AsSystem()
				} else {
					ap.As(tc.actual)
				}
				ap.Perform()
			})

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ok).To(Equal(tc.want))
		})
	}
}

func TestMatchFieldMismatch(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	m := CallAction(&childAction{}).With(Fields{"Name": Equal("expected")})

	ok, err := m.Match(func() {
		ba.New(t.Context(), &childAction{Name: "actual"}).Perform()
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestMatchActionNotCalled(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	// A different action runs; the expected one is never called.
	m := CallAction(&childAction{})

	ok, err := m.Match(func() {
		ba.New(t.Context(), &otherAction{}).Perform()
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestMatchRejectsNonFunctionActual(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	_, err := CallAction(&childAction{}).Match("not a function")
	g.Expect(err).To(HaveOccurred())
}

func TestCompositeMatcher(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	m := NewCompositeMatcher(
		CallAction(&childAction{}).AsSystem(),
		CallAction(&otherAction{}),
	)

	ok, err := m.Match(func() {
		ba.New(t.Context(), &childAction{}).AsSystem().Perform()
		ba.New(t.Context(), &otherAction{}).Perform()
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
}

// TestCallActionEndToEnd exercises the matcher the way callers actually use
// it — through Gomega's Expect/To.
func TestCallActionEndToEnd(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	g.Expect(func() {
		ba.New(t.Context(), &childAction{Name: "e2e"}).AsSystem().Perform()
	}).To(CallAction(&childAction{}).
		AsSystem().
		With(Fields{"Name": Equal("e2e")}).
		ViaPerform())
}

func TestCallActionEndToEndNegated(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	g.Expect(func() {
		ba.New(t.Context(), &otherAction{}).Perform()
	}).ShouldNot(CallAction(&childAction{}))
}

func TestMatchExtraCallDoesNotBreakConfiguredMatch(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	// Only one call is configured; a second invocation of the same action
	// exercises the tracker's unexpected-call path (an internal flag with no
	// public observable), and the configured call must still match.
	m := CallAction(&childAction{}).AsSystem()

	ok, err := m.Match(func() {
		ba.New(t.Context(), &childAction{}).AsSystem().Perform()
		ba.New(t.Context(), &childAction{}).AsSystem().Perform()
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
}

func TestFailureMessages(t *testing.T) {
	g := NewWithT(t)

	m := CallAction(&childAction{}).AsSystem()
	g.Expect(m.FailureMessage(nil)).ToNot(BeEmpty())
	g.Expect(m.NegatedFailureMessage(nil)).ToNot(BeEmpty())

	composite := NewCompositeMatcher(CallAction(&childAction{}), CallAction(&otherAction{}))
	g.Expect(composite.FailureMessage(func() {})).ToNot(BeEmpty())
	g.Expect(composite.NegatedFailureMessage(func() {})).ToNot(BeEmpty())
}

// TestMatchRestoresPreviousTracker is a regression test for the global
// ba.CallTracker leak: a match that records a call must not bleed into a
// subsequent match. Before the save/restore fix the second match below saw
// the first match's recorded childAction call and wrongly succeeded.
func TestMatchRestoresPreviousTracker(t *testing.T) {
	g := NewWithT(t)
	resetTracker(t)

	g.Expect(ba.CallTracker).To(BeNil())

	first := CallAction(&childAction{})
	ok, err := first.Match(func() {
		ba.New(t.Context(), &childAction{}).Perform()
	})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())

	// The global tracker is restored to its prior (nil) value.
	g.Expect(ba.CallTracker).To(BeNil())

	// A second match where childAction is NOT called must not match.
	second := CallAction(&childAction{})
	ok, err = second.Match(func() {
		ba.New(t.Context(), &otherAction{}).Perform()
	})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestCompositeRejectsNonFunctionActual(t *testing.T) {
	g := NewWithT(t)

	m := NewCompositeMatcher(CallAction(&childAction{}), CallAction(&otherAction{}))
	_, err := m.Match("not a function")
	g.Expect(err).To(HaveOccurred())
}
