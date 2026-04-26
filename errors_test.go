package ba_test

import (
	"github.com/bioform/ba"
	"github.com/bioform/ba/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Action Errors", func() {
	var mockAction *mocks.Action
	var performer any

	BeforeEach(func() {
		performer = "test performer"
		mockAction = &mocks.Action{}
		mockAction.On("Performer").Return(performer)
	})

	Describe("AuthorizationError", func() {
		It("should return the correct error message and performer", func() {
			err := ba.NewAuthorizationError(mockAction)
			Expect(err.Error()).To(Equal("authorization: action: *mocks.Action, performer: test performer, error: <nil>"))
		})
	})

	Describe("DisabledError", func() {
		It("should return the correct error message, performer, and error map", func() {
			errs := ba.ErrorMap{"feature": "disabled"}
			err := ba.NewDisabledError(mockAction, errs)
			Expect(err.Error()).To(Equal("not enabled: action: *mocks.Action, performer: test performer, error: map[feature:disabled]"))
			Expect(err.ActionError.Unwrap()).To(Equal(errs))
			Expect(err.Cause()).To(Equal(errs))
		})
	})

	Describe("ValidationError", func() {
		It("should return the correct error message, performer, and error map", func() {
			errs := ba.ErrorMap{"field": "invalid"}
			err := ba.NewValidationError(mockAction, errs)
			Expect(err.Error()).To(Equal("validation failed: action: *mocks.Action, performer: test performer, error: map[field:invalid]"))
			Expect(err.ActionError.Unwrap()).To(Equal(errs))
			Expect(err.Cause()).To(Equal(errs))
		})
	})
})
