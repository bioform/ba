package dummy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bioform/ba/dummy"
	. "github.com/onsi/gomega"
)

func TestTransactionProviderRunsLambda(t *testing.T) {
	g := NewWithT(t)

	called := false
	err := dummy.TransactionProvider{}.Transaction(t.Context(), func(context.Context) error {
		called = true
		return nil
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(called).To(BeTrue())
}

func TestTransactionProviderPropagatesLambdaError(t *testing.T) {
	g := NewWithT(t)

	want := errors.New("boom")
	err := dummy.TransactionProvider{}.Transaction(t.Context(), func(context.Context) error {
		return want
	})

	g.Expect(err).To(MatchError(want))
}

func TestFailingTransactionProvider(t *testing.T) {
	commitErr := errors.New("commit failed")
	lambdaErr := errors.New("lambda failed")

	cases := []struct {
		name      string
		lambdaErr error
		want      error
	}{
		// Lambda succeeds, but the transaction is still reported as failed —
		// models a commit failure so callers can assert AfterCommit is skipped.
		{"commit fails after lambda succeeds", nil, commitErr},
		// A lambda error takes precedence over the commit error.
		{"lambda error wins", lambdaErr, lambdaErr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			called := false
			err := dummy.FailingTransactionProvider{Err: commitErr}.Transaction(t.Context(), func(context.Context) error {
				called = true
				return tc.lambdaErr
			})

			g.Expect(called).To(BeTrue())
			g.Expect(err).To(MatchError(tc.want))
		})
	}
}
