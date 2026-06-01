// Package attr provides Type[T], a generic wrapper that distinguishes an
// unset value from a zero-valued one (e.g. int(0) vs. unset). Use it for
// optional action attributes; pair with attr.Required[T] as a govalidator
// rule for "set and non-empty" validation.
package attr

import (
	"fmt"
	"reflect"
)

// Type is a generic type that holds a value and a flag indicating if the value is set.
type Type[T any] struct {
	val   T
	isSet bool
}

// Value creates a new Type instance with the given value.
// The isSet flag is set to true.
func Value[T any](v T) Type[T] {
	return Type[T]{val: v, isSet: true}
}

func (v Type[T]) Val() T {
	return v.val
}

func (v Type[T]) IsSet() bool {
	return v.isSet
}

func (v Type[T]) String() string {
	return fmt.Sprintf("%v", v.val)
}

// This is a custom validation rule for the github.com/rezakhademix/govalidator/v2 package.
//
// Required checks if the given attribute value is set and non-empty/non-nil.
// An unset value is never required-satisfied. For a set value:
//   - strings, slices, maps, and arrays must be non-empty;
//   - pointers, channels, and functions must be non-nil;
//   - any other type (numbers, bools, structs, …) counts as present once set.
//
// The check uses the dynamic value, so a set-but-typed-nil pointer (e.g.
// attr.Value((*T)(nil))) or a nil slice/map is correctly reported as not
// satisfied — the typed nil does not masquerade as present.
//
// Parameters:
//   - v: The attribute value to be checked.
//
// Returns:
//   - bool: True if the attribute value is set and non-empty/non-nil, false otherwise.
func Required[T any](v Type[T]) bool {
	if !v.isSet {
		return false
	}

	switch rv := reflect.ValueOf(v.val); rv.Kind() {
	case reflect.Invalid: // a set-but-nil interface value
		return false
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() > 0
	case reflect.Pointer, reflect.Chan, reflect.Func:
		return !rv.IsNil()
	default: // numbers, bools, structs, … — present once set
		return true
	}
}
