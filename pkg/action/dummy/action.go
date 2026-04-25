package dummy

import (
	"github.com/bioform/ba/pkg/action"
	"github.com/bioform/ba/pkg/action/mocks"
	"github.com/stretchr/testify/mock"
)

type Action struct {
	mocks.Action
	performer action.Performer
}

// It is required to prevent Recursion During Stringification
func (a *Action) Performer() action.Performer {
	return nil
}

func (a *Action) SetPerformer(performer action.Performer) {
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
