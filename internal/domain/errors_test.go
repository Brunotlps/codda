package domain_test

import (
	"fmt"
	"testing"

	"github.com/Brunotlps/codda/internal/domain"
)

func TestIsValidationError(t *testing.T) {
	for _, err := range domain.ValidationErrors() {
		t.Run(err.Error(), func(t *testing.T) {
			if !domain.IsValidationError(err) {
				t.Errorf("IsValidationError(%v) = false, want true", err)
			}

			wrapped := fmt.Errorf("wrapped: %w", err)
			if !domain.IsValidationError(wrapped) {
				t.Errorf("IsValidationError(wrapped %v) = false, want true", err)
			}
		})
	}
}

func TestIsValidationErrorRejectsNonValidationErrors(t *testing.T) {
	if domain.IsValidationError(domain.ErrInvalidStatusTransition) {
		t.Errorf("IsValidationError(ErrInvalidStatusTransition) = true, want false")
	}
}

func TestValidationErrorsReturnsCopy(t *testing.T) {
	errs := domain.ValidationErrors()
	if len(errs) == 0 {
		t.Fatal("ValidationErrors() returned empty slice")
	}

	errs[0] = nil

	again := domain.ValidationErrors()
	if again[0] == nil {
		t.Errorf("ValidationErrors() returned mutable internal slice")
	}
}
