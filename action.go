// Package ba implements the Business Action pattern — a small, generic Go
// library for structuring business operations as actions: a unit of work
// with built-in authorization, validation, transaction handling, and
// after-commit callbacks.
//
// The entry point is ba.New(ctx, action).Perform(), which runs:
//
//  1. IsAllowed   — authorization (cached)
//  2. IsEnabled   — feature flag / subject-state precondition (cached)
//  3. IsValid     — validation
//  4. Perform     — business logic, wrapped in TransactionProvider.Transaction
//  5. AfterCommit — callbacks fired only if the transaction commits
//
// This file defines the Action interface every action satisfies and the
// Performer / SystemPerformer types that identify who is running the action.
package ba

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
