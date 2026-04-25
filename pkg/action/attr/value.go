package attr

import "fmt"

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
// It supports various types including string, []byte, []rune, and []int.
// For other types, it checks if the value is set and non-nil.
//
// Parameters:
//   - v: The attribute value to be checked.
//
// Returns:
//   - bool: True if the attribute value is set and non-empty/non-nil, false otherwise.
func Required[T any](v Type[T]) bool {
	switch val := any(v.val).(type) {
	case string:
		return v.isSet && val != ""
	case []byte:
		return v.isSet && len(val) > 0
	case []rune:
		return v.isSet && len(val) > 0
	case []int:
		return v.isSet && len(val) > 0
	default:
		return v.isSet && val != nil
	}
}
