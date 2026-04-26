package main_test

import (
	"context"
	"testing"

	main "github.com/bioform/ba/examples/04_nested_transactions"

	"github.com/bioform/ba"
	"github.com/bioform/ba/examples/pkg/api"
	"github.com/bioform/ba/examples/pkg/database"

	"github.com/bioform/ba/attr"
	. "github.com/bioform/ba/matcher"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestActionA(t *testing.T) {
	RegisterTestingT(t)

	ctx := api.New(database.Default()).AddTo(context.Background())

	// ActionB is stubbed (no AndCallOriginal): the matcher verifies the call
	// shape — ActionA invoked ActionB as system with these attrs — without
	// running ActionB's body. The savepoint-rollback behavior of the real
	// ActionB is exercised by `go run ./examples/04_nested_transactions`.
	Expect(func() {
		ok, err := ba.New(ctx, &main.ActionA{
			AttrA: attr.Value("Hello, World!!!!"),
		}).Perform()

		if err != nil || !ok {
			t.Errorf("Error performing action: %v", err)
		}
	}).To(CallAction(&main.ActionB{}).
		AsSystem().
		With(Fields{
			"AttrB":  Equal(attr.Value(123)),
			"AttrB2": Equal("some string"),
		}).
		ViaPerform())
}
