package error

import (
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func DiagnosticsFromError(err error) diag.Diagnostics {
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
