package error

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
)

func TestDiagnosticsError(t *testing.T) {
	t.Run(
		"it returns an empty Diagnostics if the error is nil",
		func(t *testing.T) {
			result := Diagnostics(nil)
			assert.Empty(t, result)
		},
	)
	t.Run(
		"it returns a Diagnostics containing a single error if the error is not nil",
		func(t *testing.T) {
			result := Diagnostics(fmt.Errorf("Test error"))
			assert.True(t, result.HasError())
		},
	)
	t.Run(
		"it returns a Diagnostics containing all of the errors if the error is a multierror",
		func(t *testing.T) {
			var errors *multierror.Error
			errors = multierror.Append(errors, fmt.Errorf("Test error 1"))
			errors = multierror.Append(errors, fmt.Errorf("Test error 2"))

			result := Diagnostics(errors)
			assert.Equal(t, result.ErrorsCount(), 2)
		},
	)
	t.Run(
		"it returns a Diagnostics containing all of the errors if the error is a multierror.ErrorNil()",
		func(t *testing.T) {
			var errors *multierror.Error
			errors = multierror.Append(errors, fmt.Errorf("Test error 1"))
			errors = multierror.Append(errors, fmt.Errorf("Test error 2"))

			result := Diagnostics(errors.ErrorOrNil())
			assert.Equal(t, result.ErrorsCount(), 2)
		},
	)
}
