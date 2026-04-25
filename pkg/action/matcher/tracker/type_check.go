package tracker

import (
	"reflect"

	"github.com/bioform/ba/pkg/action"
)

func typeOf(action action.Action) reflect.Type {
	actionType := reflect.TypeOf(action)

	if actionType.Kind() == reflect.Ptr {
		actionType = actionType.Elem()
	}

	return actionType
}
