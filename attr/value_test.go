package attr_test

import (
	"testing"

	"github.com/bioform/ba/attr"
	. "github.com/onsi/gomega"
)

func TestValueSetsAndReports(t *testing.T) {
	g := NewWithT(t)

	v := attr.Value("hello")
	g.Expect(v.IsSet()).To(BeTrue())
	g.Expect(v.Val()).To(Equal("hello"))
	g.Expect(v.String()).To(Equal("hello"))
}

func TestZeroValueIsUnset(t *testing.T) {
	g := NewWithT(t)

	// A freshly declared Type[T] is distinguishable from an explicitly-set
	// zero value: IsSet reports false even though Val is the zero value.
	var unset attr.Type[int]
	g.Expect(unset.IsSet()).To(BeFalse())
	g.Expect(unset.Val()).To(Equal(0))

	set := attr.Value(0)
	g.Expect(set.IsSet()).To(BeTrue())
	g.Expect(set.Val()).To(Equal(0))
}

func TestString(t *testing.T) {
	cases := []struct {
		name string
		got  func() string
		want string
	}{
		{"string", func() string { return attr.Value("hello").String() }, "hello"},
		{"int", func() string { return attr.Value(123).String() }, "123"},
		{"bool", func() string { return attr.Value(true).String() }, "true"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			NewWithT(t).Expect(tc.got()).To(Equal(tc.want))
		})
	}
}

func TestRequired(t *testing.T) {
	// Required is generic, so each case runs it against a different T inside
	// the subtest (via a closure) and compares the resulting bool — letting
	// every branch share one table while keeping the call under the subtest.
	cases := []struct {
		name string
		run  func() bool
		want bool
	}{
		{"string set non-empty", func() bool { return attr.Required(attr.Value("x")) }, true},
		{"string set empty", func() bool { return attr.Required(attr.Value("")) }, false},
		{"string unset", func() bool { return attr.Required(attr.Type[string]{}) }, false},

		{"[]byte non-empty", func() bool { return attr.Required(attr.Value([]byte{1})) }, true},
		{"[]byte empty", func() bool { return attr.Required(attr.Value([]byte{})) }, false},
		{"[]byte unset", func() bool { return attr.Required(attr.Type[[]byte]{}) }, false},

		{"[]rune non-empty", func() bool { return attr.Required(attr.Value([]rune{'a'})) }, true},
		{"[]rune empty", func() bool { return attr.Required(attr.Value([]rune{})) }, false},
		{"[]rune unset", func() bool { return attr.Required(attr.Type[[]rune]{}) }, false},

		{"[]int non-empty", func() bool { return attr.Required(attr.Value([]int{1})) }, true},
		{"[]int empty", func() bool { return attr.Required(attr.Value([]int{})) }, false},
		{"[]int unset", func() bool { return attr.Required(attr.Type[[]int]{}) }, false},

		// Default branch: any set value counts, even the zero value.
		{"int zero but set", func() bool { return attr.Required(attr.Value(0)) }, true},
		{"int non-zero", func() bool { return attr.Required(attr.Value(42)) }, true},
		{"int unset", func() bool { return attr.Required(attr.Type[int]{}) }, false},

		// Sharp edges of the default branch: a *set* but nil/empty value whose
		// type has no dedicated branch still counts as required, because
		// any(typedNil) is a non-nil interface. These lock in current behavior
		// (and document that the "non-nil" doc comment overstates the contract).
		{"nil pointer, set", func() bool { return attr.Required(attr.Value((*int)(nil))) }, true},
		{"nil non-special slice, set", func() bool { return attr.Required(attr.Value([]string(nil))) }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			NewWithT(t).Expect(tc.run()).To(Equal(tc.want))
		})
	}
}
