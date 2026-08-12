package error

import (
	"errors"

	managementv3 "github.com/auth0/go-auth0/v3/management"
)

// v3ForbiddenErrorCode extracts the errorCode field from a v3 SDK ForbiddenError.
// Returns an empty string if the error is not a ForbiddenError or if the
// Body does not contain a valid errorCode field.
func v3ForbiddenErrorCode(err error) string {
	var fe *managementv3.ForbiddenError
	if !errors.As(err, &fe) {
		return ""
	}

	body, ok := fe.Body.(map[string]interface{})
	if !ok {
		return ""
	}

	code, _ := body["errorCode"].(string)
	return code
}

// IsInsufficientEntitlement checks if the error is a v3 SDK ForbiddenError
// with errorCode "insufficient_entitlement", indicating the tenant lacks
// the required entitlement for the requested feature.
func IsInsufficientEntitlement(err error) bool {
	return v3ForbiddenErrorCode(err) == "insufficient_entitlement"
}

// IsInsufficientScope checks if the error is a v3 SDK ForbiddenError
// with errorCode "insufficient_scope", indicating the access token lacks
// the required scope for the requested operation.
func IsInsufficientScope(err error) bool {
	return v3ForbiddenErrorCode(err) == "insufficient_scope"
}
