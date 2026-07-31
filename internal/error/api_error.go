package error

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/auth0/go-auth0/v2/management/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// HandleAPIError handles the API error.
// If the error code is a 404 it triggers a resource deletion.
func HandleAPIError(data *schema.ResourceData, err error) error {
	if IsStatusNotFound(err) {
		data.SetId("")
		return nil
	}

	return err
}

// HandleReadAPIError handles an API error returned from a resource's read operation.
// On a 404 it removes the resource from state (like HandleAPIError) but also emits a
// warning explaining that the resource no longer exists in Auth0 and how to proceed,
// so the removal is not silent. All other errors are returned as-is.
//
// The resourceType is the Terraform type of the resource being read (e.g. "auth0_action")
// and is used to render an actionable `terraform state rm` command in the warning.
func HandleReadAPIError(resourceType string, data *schema.ResourceData, err error) diag.Diagnostics {
	if IsStatusNotFound(err) {
		return RemoveFromStateWithWarning(resourceType, data, "the API returned 404")
	}

	return diag.FromErr(err)
}

// RemoveFromStateWithWarning removes the resource from the Terraform state and returns a
// warning explaining that it no longer exists in Auth0 and how to proceed, so the removal
// is never silent. Use this for reads that detect a vanished resource without an API error,
// for example when the resource is a member of a list that no longer contains it.
//
// The resourceType is the Terraform type of the resource being read (e.g. "auth0_action")
// and is used to render an actionable `terraform state rm` command in the warning. The
// reason explains how the absence was detected (e.g. "the API returned 404").
func RemoveFromStateWithWarning(resourceType string, data *schema.ResourceData, reason string) diag.Diagnostics {
	id := data.Id()
	data.SetId("")

	return diag.Diagnostics{{
		Severity: diag.Warning,
		Summary:  "Resource not found, removed from state",
		Detail: fmt.Sprintf(
			"The %s resource with ID %q was not found in Auth0 (%s) and has "+
				"been removed from the Terraform state automatically. It was most likely deleted "+
				"outside of Terraform.\n\n"+
				"If this was expected, no action is needed. The next plan will reconcile the "+
				"state. To recreate the resource, run `terraform apply`. To drop it from state "+
				"manually instead, run `terraform state rm %s.<name>`, using the resource name "+
				"shown in the address above.",
			resourceType, id, reason, resourceType,
		),
	}}
}

// IsStatusNotFound checks to see if the error from the Auth0 Management API is a 404.
// It understands both the v1 SDK error type, which exposes the status code through the
// management.Error interface, and the v2 SDK error types, which wrap a *core.APIError
// carrying the status code (e.g. *management.NotFoundError).
func IsStatusNotFound(err error) bool {
	if err == nil {
		return false
	}

	// V1 SDK: errors implement management.Error with a Status() method.
	var mErr management.Error
	if errors.As(err, &mErr) && mErr.Status() == http.StatusNotFound {
		return true
	}

	// V2 SDK: errors embed *core.APIError, which holds the status code in a field.
	var apiErr *core.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return true
	}

	return false
}
