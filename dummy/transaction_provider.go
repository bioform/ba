package dummy

import (
	"context"

	"github.com/bioform/ba"
)

var (
	_ ba.TransactionProvider = TransactionProvider{}
	_ ba.TransactionProvider = FailingTransactionProvider{}
)

// TransactionProvider is a no-op ba.TransactionProvider for unit tests: it
// runs the lambda directly on the given context, with no real transaction.
// Use it when an action under test needs a TransactionProvider but the test
// doesn't care about transactional semantics, avoiding the testify-mock
// lambda boilerplate.
type TransactionProvider struct{}

func (TransactionProvider) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// FailingTransactionProvider runs the lambda but always reports the
// transaction as failed by returning Err — even when the lambda itself
// succeeds. It models a commit failure, letting tests assert that
// after-commit callbacks are skipped when the transaction does not commit.
type FailingTransactionProvider struct {
	Err error
}

func (p FailingTransactionProvider) Transaction(ctx context.Context, fn func(context.Context) error) error {
	if err := fn(ctx); err != nil {
		return err
	}
	return p.Err
}
