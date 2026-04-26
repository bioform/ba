// Package dummy provides a testify/mock-backed Action implementation for
// unit tests that need a stand-in Action without writing custom mocks.
// Use dummy.NewAction(t) to obtain an instance whose expectations are
// asserted automatically on test cleanup.
package dummy

import (
	"github.com/bioform/ba"
	"github.com/bioform/ba/mocks"
	"github.com/stretchr/testify/mock"
)

type Action struct {
	mocks.Action
	performer ba.Performer
}

// It is required to prevent Recursion During Stringification
func (a *Action) Performer() ba.Performer {
	return nil
}

func (a *Action) SetPerformer(performer ba.Performer) {
	a.performer = performer
}

func NewAction(t interface {
	mock.TestingT
	Cleanup(func())
}) *Action {
	mock := &Action{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
