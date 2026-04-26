// TransactionProvider — the single interface a host application must
// implement to plug ba into its persistence layer. Implementations should
// (1) start a transaction, (2) call executeInTransaction with a context
// that carries the transactional handle, and (3) commit on nil error or
// roll back otherwise. Nested calls should produce savepoints.
package ba

import (
	"context"
)

type TransactionProvider interface {
	Transaction(currentContext context.Context, executeInTransaction func(newContext context.Context) error) error
}
