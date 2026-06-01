package option_test

import (
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/matcher/option"
	. "github.com/onsi/gomega"
)

func TestTrackOptionsBuilders(t *testing.T) {
	g := NewWithT(t)

	g.Expect(option.New()).To(Equal(option.TrackOptions{}))
	g.Expect(option.CallOriginal().CallOriginal).To(BeTrue())
	g.Expect(option.With(option.Try).Method).To(Equal(option.Try))
	g.Expect(option.New().AndCallOriginal().CallOriginal).To(BeTrue())
	g.Expect(option.New().With(option.Perform).Method).To(Equal(option.Perform))
}

func TestMethodString(t *testing.T) {
	cases := []struct {
		method option.Method
		want   string
	}{
		{option.Perform, "Perform"},
		{option.Try, "Try"},
		{option.None, "None"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			NewWithT(t).Expect(tc.method.String()).To(Equal(tc.want))
		})
	}
}

func TestGetMethod(t *testing.T) {
	cases := []struct {
		name string
		opts []ba.Option
		want option.Method
	}{
		{"no options means Perform", nil, option.Perform},
		{"NopIfDisabled means Try", []ba.Option{ba.NopIfDisabled}, option.Try},
		{"other options still Perform", []ba.Option{ba.SkipCache}, option.Perform},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			NewWithT(t).Expect(option.GetMethod(tc.opts)).To(Equal(tc.want))
		})
	}
}
