package http

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Brunotlps/codda/internal/domain"
)

func TestHTTPStatusForErrorMapsAllDomainValidationErrors(t *testing.T) {
	for _, err := range domain.ValidationErrors() {
		t.Run(err.Error(), func(t *testing.T) {
			status, code, message := httpStatusForError(fmt.Errorf("wrapped: %w", err))

			if status != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
			}
			if code != "validation_error" {
				t.Errorf("code = %q, want %q", code, "validation_error")
			}
			if message == "" {
				t.Errorf("message is empty")
			}
		})
	}
}
