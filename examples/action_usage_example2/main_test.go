package main_test

import (
	"context"
	"testing"

	main "github.com/bioform/ba/examples/action_usage_example2"

	"github.com/bioform/ba/pkg/action"
	"github.com/bioform/ba/examples/pkg/api"
	"github.com/bioform/ba/examples/pkg/database"

	"github.com/bioform/ba/pkg/action/attr"
	. "github.com/bioform/ba/pkg/action/matcher"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestActionA(t *testing.T) {
	RegisterTestingT(t)

	ctx := api.New(database.Default()).AddTo(context.Background())

	Expect(func() {
		// Code that triggers the action
		ok, err := action.New(ctx, &main.ActionA{
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
		ViaPerform().
		AndCallOriginal())
}
