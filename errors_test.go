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

	// These tests assert the stable, meaningful parts of each error — the
	// phase prefix, the performer, and the structured cause — rather than the
	// exact fmt layout or the concrete mock type name, so formatting tweaks or
	// a mock rename don't break them.

	Describe("AuthorizationError", func() {
		It("should report the phase and performer", func() {
			err := ba.NewAuthorizationError(mockAction)
			Expect(err.Error()).To(HavePrefix("authorization:"))
			Expect(err.Error()).To(ContainSubstring("test performer"))
		})
	})

	Describe("DisabledError", func() {
		It("should report the phase, performer, and error map", func() {
			errs := ba.ErrorMap{"feature": "disabled"}
			err := ba.NewDisabledError(mockAction, errs)

			Expect(err.Error()).To(HavePrefix("not enabled:"))
			Expect(err.Error()).To(ContainSubstring("test performer"))
			Expect(err.Cause()).To(Equal(errs))
			Expect(err.ActionError.Unwrap()).To(Equal(errs))
		})
	})

	Describe("ValidationError", func() {
		It("should report the phase, performer, and error map", func() {
			errs := ba.ErrorMap{"field": "invalid"}
			err := ba.NewValidationError(mockAction, errs)

			Expect(err.Error()).To(HavePrefix("validation failed:"))
			Expect(err.Error()).To(ContainSubstring("test performer"))
			Expect(err.Cause()).To(Equal(errs))
			Expect(err.ActionError.Unwrap()).To(Equal(errs))
		})
	})
})
