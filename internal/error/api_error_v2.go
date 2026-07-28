package error

import (
	"errors"

	managementv2 "github.com/auth0/go-auth0/v2/management"
)

// v2ErrorCode extracts the errorCode field from a v2 SDK ForbiddenError.
// Returns an empty string if the error is not a ForbiddenError or if the
// Body does not contain a valid errorCode field.
func v2ErrorCode(err error) string {
	var fe *managementv2.ForbiddenError
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

// IsInsufficientEntitlement checks if the error is a v2 SDK ForbiddenError
// with errorCode "insufficient_entitlement", indicating the tenant lacks
// the required entitlement for the requested feature.
func IsInsufficientEntitlement(err error) bool {
	return v2ErrorCode(err) == "insufficient_entitlement"
}

// IsInsufficientScope checks if the error is a v2 SDK ForbiddenError
// with errorCode "insufficient_scope", indicating the access token lacks
// the required scope for the requested operation.
func IsInsufficientScope(err error) bool {
	return v2ErrorCode(err) == "insufficient_scope"
}
