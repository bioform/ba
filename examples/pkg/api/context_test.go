package api

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
)

func TestFromReturnsErrNoAPI(t *testing.T) {
	g := NewWithT(t)

	_, err := From(t.Context())
	g.Expect(err).To(MatchError(ErrNoAPI))
}

func TestFromReturnsErrInvalidAPI(t *testing.T) {
	g := NewWithT(t)

	// Something other than *api stored under the key.
	ctx := context.WithValue(t.Context(), apiKey, "not an api")
	_, err := From(ctx)
	g.Expect(err).To(MatchError(ErrInvalidAPI))
}

func TestFromReturnsStoredNilAPI(t *testing.T) {
	g := NewWithT(t)

	// A typed-nil *api stored under the key is a non-nil interface, so From's
	// nil and type-assertion checks both pass and it returns (nil, nil). This
	// locks in that edge behavior.
	ctx := context.WithValue(t.Context(), apiKey, (*api)(nil))
	got, err := From(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(got).To(BeNil())
}

func TestAddToRoundTrip(t *testing.T) {
	g := NewWithT(t)

	a := New(nil)
	got, err := From(a.AddTo(t.Context()))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(got).To(BeIdenticalTo(a))
}
