package error

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// HandleAPIError handles an API error. If it is a 404 Not Found,
// the resource is deleted.
func HandleAPIError(ctx context.Context, responseState *tfsdk.State, err error) diag.Diagnostics {
	result := diag.Diagnostics{}
	if err == nil {
		return result
	}
	if IsStatusNotFound(err) {
		responseState.RemoveResource(ctx)
		result.AddWarning("Resource missing", err.Error())
	} else {
		result.Append(Diagnostics(err)...)
	}

	return result
}

// IsStatusNotFound checks to see if the error from the Auth0 Management API is a 404.
func IsStatusNotFound(err error) bool {
	if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
		return true
	}

	return false
}
