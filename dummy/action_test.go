package dummy_test

import (
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/dummy"
	. "github.com/onsi/gomega"
)

func TestNewActionImplementsAction(t *testing.T) {
	g := NewWithT(t)

	a := dummy.NewAction(t)
	var _ ba.Action = a
	g.Expect(a).ToNot(BeNil())
}

// TestPerformerOverridePreventsRecursion guards the dummy.Action.Performer
// override. The underlying testify mock stringifies its expected calls; if
// Performer were left to the mock it could recurse during stringification.
// The override always returns nil regardless of what SetPerformer recorded.
func TestPerformerOverride(t *testing.T) {
	g := NewWithT(t)

	a := dummy.NewAction(t)
	g.Expect(a.Performer()).To(BeNil())

	a.SetPerformer("someone")
	g.Expect(a.Performer()).To(BeNil())
}
