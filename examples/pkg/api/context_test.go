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

func TestAddToRoundTrip(t *testing.T) {
	g := NewWithT(t)

	a := New(nil)
	got, err := From(a.AddTo(t.Context()))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(got).To(BeIdenticalTo(a))
}
