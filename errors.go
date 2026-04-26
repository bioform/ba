// Error hierarchy for the ba package. Every error wraps ActionError, which
// carries the failing action and its performer for context. The specific
// subtypes (AuthorizationError, DisabledError, ValidationError,
// CallbackError) tell callers exactly which lifecycle phase failed.
package ba

import (
	"fmt"
)

type ErrorMapWrapper interface {
	Cause() ErrorMap
	Add(key, value string)
}
type ErrorMap map[string]string

type ActionError struct {
	action Action
	error
	stack []byte
}

type CallbackError struct {
	ActionError
}

type AuthorizationError struct {
	ActionError
}

type DisabledError struct {
	ActionError
	ErrorMapWrapper
}

type ValidationError struct {
	ActionError
	ErrorMapWrapper
}

func (e ErrorMap) Error() string {
	return fmt.Sprintf("%v", map[string]string(e))
}

func NewActionError(action Action, err error) *ActionError {
	return &ActionError{action: action, error: err}
}

func NewCallbackError(action Action, err error) *CallbackError {
	return &CallbackError{ActionError{action: action, error: err}}
}

func NewCallbackErrorWithStack(action Action, err error, stack []byte) *CallbackError {
	return &CallbackError{ActionError{action: action, error: err, stack: stack}}
}

func NewAuthorizationError(action Action) *AuthorizationError {
	return &AuthorizationError{ActionError: ActionError{action: action}}
}

func NewDisabledError(action Action, errs ErrorMap) *DisabledError {
	return &DisabledError{
		ActionError: ActionError{action: action, error: errs},
	}
}

func NewValidationError(action Action, errs ErrorMap) *ValidationError {
	return &ValidationError{
		ActionError: ActionError{action: action, error: errs},
	}
}

func (e ActionError) Unwrap() error {
	return e.error
}

func (e ActionError) Cause() error {
	return e.error
}

func (e CallbackError) Unwrap() error {
	return e.ActionError
}

func (e AuthorizationError) Unwrap() error {
	return e.ActionError
}

func (e DisabledError) Unwrap() error {
	return e.ActionError
}

func (e ValidationError) Unwrap() error {
	return e.ActionError
}

func (e ActionError) Error() string {
	if len(e.stack) > 0 {
		return fmt.Sprintf("action: %T, performer: %v, error: %v\n%s", e.action, e.action.Performer(), e.error, e.stack)
	}
	return fmt.Sprintf("action: %T, performer: %v, error: %v", e.action, e.action.Performer(), e.error)
}

func (e CallbackError) Error() string {
	return fmt.Sprintf("callback: %v", e.ActionError)
}

func (e AuthorizationError) Error() string {
	return fmt.Sprintf("authorization: %v", e.ActionError)
}

func (e DisabledError) Error() string {
	return fmt.Sprintf("not enabled: %v", e.ActionError)
}

func (e DisabledError) Cause() ErrorMap {
	return e.error.(ErrorMap)
}

func (e DisabledError) Add(key, value string) {
	e.error.(ErrorMap)[key] = value
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.ActionError)
}

func (e ValidationError) Cause() ErrorMap {
	return e.error.(ErrorMap)
}

func (e ValidationError) Add(key, value string) {
	e.error.(ErrorMap)[key] = value
}
