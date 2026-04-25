package action

import (
	"context"
)

var SystemPerformer systemPerformer = systemPerformer{}

type Performer any

type systemPerformer struct{}

type Action interface {
	Init() // You can ovveride this method in your action. Usually used to set default values for your action attributes.
	SetContext(context.Context)
	Context() context.Context
	ContextUpdated() // called after the context is updated
	TransactionProvider() TransactionProvider
	Performer() Performer
	SetPerformer(Performer)
	Perform() error // This is the main method that should be implemented by the action.
	IsAllowed() (bool, error)
	IsEnabled() (bool, error)
	IsValid() (bool, error)
	AfterCommitCallback() AfterCommitCallback
	ErrorHandler(error) error
}

type AfterCommitCallback func(context.Context, Action) error

func (systemPerformer) String() string {
	return "system"
}
