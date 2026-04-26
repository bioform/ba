// Reflection helper: typeOf returns the (dereferenced) reflect.Type of an
// Action so the tracker can key its records by concrete action type even
// when callers pass *T pointers.
package tracker

import (
	"reflect"

	"github.com/bioform/ba"
)

func typeOf(action ba.Action) reflect.Type {
	actionType := reflect.TypeOf(action)

	if actionType.Kind() == reflect.Ptr {
		actionType = actionType.Elem()
	}

	return actionType
}
