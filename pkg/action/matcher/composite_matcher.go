package matcher

import (
	"errors"
	"fmt"
	"strings"

	"github.com/onsi/gomega/types"
)

type compositeMatcher struct {
	first  types.GomegaMatcher
	second types.GomegaMatcher
}

func NewCompositeMatcher(first, second types.GomegaMatcher) types.GomegaMatcher {
	return &compositeMatcher{
		first:  first,
		second: second,
	}
}

func (c *compositeMatcher) Match(actual interface{}) (bool, error) {
	actualFunc, ok := actual.(func())
	if !ok {
		return false, fmt.Errorf("CompositeMatcher expects a function as actual")
	}

	var (
		firstSuccess bool
		firstError   error
	)
	wrappedFunc := func() {
		firstSuccess, firstError = c.first.Match(actualFunc)
	}

	var (
		secondSuccess bool
		secondError   error
	)
	secondSuccess, secondError = c.second.Match(wrappedFunc)

	if firstError != nil && secondError != nil {
		return false, errors.Join(firstError, secondError)
	}
	if firstError != nil {
		return false, firstError
	}
	if secondError != nil {
		return false, secondError
	}
	return firstSuccess && secondSuccess, nil
}

func (c *compositeMatcher) FailureMessage(actual interface{}) string {
	return join(c.first.FailureMessage(actual), c.second.FailureMessage(actual))
}

func (c *compositeMatcher) NegatedFailureMessage(actual interface{}) string {
	return join(c.first.NegatedFailureMessage(actual), c.second.NegatedFailureMessage(actual))
}

func (m *compositeMatcher) And(next types.GomegaMatcher) types.GomegaMatcher {
	return NewCompositeMatcher(m, next)
}

func join(stringList ...string) string {
	var nonEmptyStrings []string
	for _, s := range stringList {
		if len(s) > 0 {
			nonEmptyStrings = append(nonEmptyStrings, s)
		}
	}
	return strings.Join(nonEmptyStrings, " and ")
}
