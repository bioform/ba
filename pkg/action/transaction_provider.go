package action

import (
	"context"
)

type TransactionProvider interface {
	Transaction(currentContext context.Context, executeInTransaction func(newContext context.Context) error) error
}
