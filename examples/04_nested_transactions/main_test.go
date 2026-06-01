package main_test

import (
	"testing"

	main "github.com/bioform/ba/examples/04_nested_transactions"

	"github.com/bioform/ba"
	"github.com/bioform/ba/examples/pkg/api"
	"github.com/bioform/ba/examples/pkg/model"

	"github.com/bioform/ba/attr"
	. "github.com/bioform/ba/matcher"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB builds a fresh in-memory database so each test is isolated from
// the package-global database.Default() and from other tests.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}

// TestNestedSavepointRollsBackInner asserts the documented nested-transaction
// contract end-to-end: ActionA commits its "outer" row while ActionB's failure
// rolls back only the inner savepoint — so exactly one user survives.
func TestNestedSavepointRollsBackInner(t *testing.T) {
	g := NewWithT(t)

	db := newTestDB(t)
	ctx := api.New(db).AddTo(t.Context())

	ok, err := ba.New(ctx, &main.ActionA{AttrA: attr.Value("Hello, World!")}).Perform()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())

	var users []model.User
	g.Expect(db.Find(&users).Error).ToNot(HaveOccurred())
	g.Expect(users).To(HaveLen(1))
	g.Expect(users[0].Name).To(Equal("outer"))
}

func TestActionA(t *testing.T) {
	g := NewWithT(t)

	ctx := api.New(newTestDB(t)).AddTo(t.Context())

	// ActionB is stubbed (no AndCallOriginal): the matcher verifies the call
	// shape — ActionA invoked ActionB as system with these attrs — without
	// running ActionB's body. The savepoint-rollback behavior of the real
	// ActionB is exercised by TestNestedSavepointRollsBackInner.
	g.Expect(func() {
		ok, err := ba.New(ctx, &main.ActionA{
			AttrA: attr.Value("Hello, World!!!!"),
		}).Perform()

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ok).To(BeTrue())
	}).To(CallAction(&main.ActionB{}).
		AsSystem().
		With(Fields{
			"AttrB":  Equal(attr.Value(123)),
			"AttrB2": Equal("some string"),
		}).
		ViaPerform())
}
