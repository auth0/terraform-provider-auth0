package error

import (
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Diagnostics converts an error to a framework Diagnostics
// If the error code is a 404 it triggers a resource deletion.
func Diagnostics(err error) diag.Diagnostics {
	result := diag.Diagnostics{}
	if err == nil {
		return result
	}
	if mErr, ok := err.(*multierror.Error); ok {
		if len(mErr.Errors) == 0 {
			return result
		}
		for _, err := range mErr.Errors {
			result.AddError("Error", err.Error())
		}
	} else {
		result.AddError("Error", err.Error())
	}

	return result
}
