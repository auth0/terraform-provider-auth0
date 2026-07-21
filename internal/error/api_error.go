package error

import (
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
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
		id := data.Id()
		data.SetId("")

		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  "Resource not found, removed from state",
			Detail: fmt.Sprintf(
				"The %s resource with ID %q was not found in Auth0 (the API returned 404) and has "+
					"been removed from the Terraform state automatically. It was most likely deleted "+
					"outside of Terraform.\n\n"+
					"If this was expected, no action is needed. The next plan will reconcile the "+
					"state. To recreate the resource, run `terraform apply`. To drop it from state "+
					"manually instead, run `terraform state rm %s.<name>`, using the resource name "+
					"shown in the address above.",
				resourceType, id, resourceType,
			),
		}}
	}

	return diag.FromErr(err)
}

// IsStatusNotFound checks to see if the error from the Auth0 Management API is a 404.
func IsStatusNotFound(err error) bool {
	if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
		return true
	}

	return false
}
