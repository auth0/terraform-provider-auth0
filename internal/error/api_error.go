package error

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/auth0/go-auth0/v3/management/core"
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

// IsInsufficientEntitlement reports whether err represents an Auth0 403 response
// for either a subscription-entitlement gate (errorCode "insufficient_entitlement")
// or a tenant feature-flag gate (errorCode "not_entitled", e.g. mfa_advanced_factor_config).
// It handles both the v1 and v3 SDK error types. Use this in Read and Update functions
// for entitlement-gated resources.
func IsInsufficientEntitlement(err error) bool {
	if err == nil {
		return false
	}

	// V1 SDK: management.Error exposes Status() and Code().
	var mErr management.Error
	if errors.As(err, &mErr) && mErr.Status() == http.StatusForbidden {
		return mErr.Code() == "insufficient_entitlement" || mErr.Code() == "not_entitled"
	}

	// V3 SDK: delegate to the existing v3-specific helper.
	return isInsufficientEntitlementV3(err)
}

// EntitlementReadConsequence is the standard consequence phrase for read operations
// blocked by a missing entitlement.
const EntitlementReadConsequence = "its current configuration could not be read"

// EntitlementUpdateConsequence is the standard consequence phrase for update operations
// blocked by a missing entitlement.
const EntitlementUpdateConsequence = "the configuration was not applied"

// EntitlementWarning returns a non-fatal warning diagnostic for an entitlement-gated
// feature. Pass the feature name, a consequence phrase (e.g. EntitlementUpdateConsequence),
// and the original error. The Detail text is tailored to the specific error code:
// "insufficient_entitlement" → subscription/add-on language; "not_entitled" → feature-flag
// language. The raw API message is intentionally NOT echoed; only the machine-generated
// errorCode is included to avoid surfacing known backend copy-paste bugs.
func EntitlementWarning(feature, consequence string, err error) diag.Diagnostic {
	code := forbiddenErrorCode(err)

	var gateDescription string
	if code == "not_entitled" {
		gateDescription = "a tenant feature flag that is not enabled"
	} else {
		gateDescription = "an add-on entitlement not present"
	}

	detail := fmt.Sprintf(
		"%s requires %s on this tenant, so %s (error code: %q). "+
			"Contact Auth0 support to enable this feature.",
		feature, gateDescription, consequence, code,
	)

	return diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  fmt.Sprintf("%s entitlement not available", feature),
		Detail:   detail,
	}
}

// forbiddenErrorCode extracts the errorCode field from a 403 API error,
// trying the v1 management.Error interface first, then the v3 ForbiddenError path.
// Returns "" if the error is nil or does not carry an errorCode.
func forbiddenErrorCode(err error) string {
	if err == nil {
		return ""
	}
	var mErr management.Error
	if errors.As(err, &mErr) && mErr.Status() == http.StatusForbidden {
		return mErr.Code()
	}
	return v3ForbiddenErrorCode(err)
}

// IsStatusNotFound checks to see if the error from the Auth0 Management API is a 404.
// It understands both the v1 SDK error type, which exposes the status code through the
// management.Error interface, and the v3 SDK error types, which wrap a *core.APIError
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

	// V3 SDK: errors embed *core.APIError, which holds the status code in a field.
	var apiErr *core.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return true
	}

	return false
}
