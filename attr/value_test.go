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
		got  string
		want string
	}{
		{"string", attr.Value("hello").String(), "hello"},
		{"int", attr.Value(123).String(), "123"},
		{"bool", attr.Value(true).String(), "true"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			NewWithT(t).Expect(tc.got).To(Equal(tc.want))
		})
	}
}

func TestRequired(t *testing.T) {
	// Required is generic, so each case evaluates it against a different T and
	// records the resulting bool — letting every branch share one table.
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"string set non-empty", attr.Required(attr.Value("x")), true},
		{"string set empty", attr.Required(attr.Value("")), false},
		{"string unset", attr.Required(attr.Type[string]{}), false},

		{"[]byte non-empty", attr.Required(attr.Value([]byte{1})), true},
		{"[]byte empty", attr.Required(attr.Value([]byte{})), false},
		{"[]byte unset", attr.Required(attr.Type[[]byte]{}), false},

		{"[]rune non-empty", attr.Required(attr.Value([]rune{'a'})), true},
		{"[]rune empty", attr.Required(attr.Value([]rune{})), false},

		{"[]int non-empty", attr.Required(attr.Value([]int{1})), true},
		{"[]int empty", attr.Required(attr.Value([]int{})), false},

		// Default branch: any set value counts, even the zero value.
		{"int zero but set", attr.Required(attr.Value(0)), true},
		{"int non-zero", attr.Required(attr.Value(42)), true},
		{"int unset", attr.Required(attr.Type[int]{}), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			NewWithT(t).Expect(tc.got).To(Equal(tc.want))
		})
	}
}
