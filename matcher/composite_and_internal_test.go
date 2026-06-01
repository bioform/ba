package matcher

import (
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/dummy"
	. "github.com/onsi/gomega"
)

// These actions are local to the internal test package; CallActionMatcher keys
// by concrete type, so each chained expectation needs a distinct type.
type andA struct{ ba.BaseAction }

func (a *andA) Perform() error                              { return nil }
func (a *andA) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

type andB struct{ ba.BaseAction }

func (a *andB) Perform() error                              { return nil }
func (a *andB) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

type andC struct{ ba.BaseAction }

func (a *andC) Perform() error                              { return nil }
func (a *andC) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

// TestCompositeAndChainsThree covers (*compositeMatcher).And, which chains a
// third expectation. It lives in the internal package because the
// compositeMatcher type returned by NewCompositeMatcher is unexported, so And
// is unreachable from an external test.
func TestCompositeAndChainsThree(t *testing.T) {
	g := NewWithT(t)
	ba.CallTracker = nil
	t.Cleanup(func() { ba.CallTracker = nil })

	base := NewCompositeMatcher(CallAction(&andA{}), CallAction(&andB{})).(*compositeMatcher)
	chained := base.And(CallAction(&andC{}))

	ok, err := chained.Match(func() {
		ba.New(t.Context(), &andA{}).Perform()
		ba.New(t.Context(), &andB{}).Perform()
		ba.New(t.Context(), &andC{}).Perform()
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())

	// Missing the third action makes the chained matcher fail.
	ba.CallTracker = nil
	ok, err = base.And(CallAction(&andC{})).Match(func() {
		ba.New(t.Context(), &andA{}).Perform()
		ba.New(t.Context(), &andB{}).Perform()
	})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}
