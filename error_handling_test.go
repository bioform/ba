package ba_test

import (
	"errors"
	"testing"

	"github.com/bioform/ba"
	"github.com/bioform/ba/dummy"
	. "github.com/onsi/gomega"
)

type emailDuplicateError struct{}

func (*emailDuplicateError) Error() string { return "email duplicate" }

// translatingAction maps its Perform error to a domain error in ErrorHandler,
// the pattern the README documents (translate gorm.ErrDuplicatedKey, etc.).
type translatingAction struct {
	ba.BaseAction
}

func (a *translatingAction) Perform() error                              { return errors.New("duplicated key") }
func (a *translatingAction) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }
func (a *translatingAction) ErrorHandler(error) error                    { return &emailDuplicateError{} }

func TestErrorHandlerTranslatesError(t *testing.T) {
	g := NewWithT(t)

	ok, err := ba.New(t.Context(), &translatingAction{}).Perform()
	g.Expect(ok).To(BeFalse())

	// The translated error is returned verbatim...
	var dup *emailDuplicateError
	g.Expect(errors.As(err, &dup)).To(BeTrue())
	// ...not wrapped in an ActionError.
	var ae *ba.ActionError
	g.Expect(errors.As(err, &ae)).To(BeFalse())
}

// wrapAction returns whatever error it is told to.
type wrapAction struct {
	ba.BaseAction

	err error
}

func (a *wrapAction) Perform() error                              { return a.err }
func (a *wrapAction) TransactionProvider() ba.TransactionProvider { return dummy.TransactionProvider{} }

func TestWrapErrorRewrapsForeignActionError(t *testing.T) {
	g := NewWithT(t)

	// The error already belongs to a *different* action.
	foreign := ba.NewAuthorizationError(&translatingAction{})
	a := &wrapAction{err: foreign}

	ok, err := ba.New(t.Context(), a).Perform()
	g.Expect(ok).To(BeFalse())

	// It is re-wrapped under the current action, preserving the original in
	// the chain.
	g.Expect(err).ToNot(BeIdenticalTo(error(foreign)))
	g.Expect(errors.Is(err, foreign)).To(BeTrue())
}

func TestWrapErrorPreservesOwnActionError(t *testing.T) {
	g := NewWithT(t)

	// The error already belongs to this action — it must pass through
	// untouched rather than being wrapped again.
	a := &wrapAction{}
	own := ba.NewAuthorizationError(a)
	a.err = own

	ok, err := ba.New(t.Context(), a).Perform()
	g.Expect(ok).To(BeFalse())

	authErr, isAuth := err.(*ba.AuthorizationError)
	g.Expect(isAuth).To(BeTrue())
	g.Expect(authErr).To(BeIdenticalTo(own))
}
