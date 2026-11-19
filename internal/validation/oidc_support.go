package validation

import (
	"errors"
	"fmt"
	"slices"
)

// OIDCSupportOptionalScopes are the optional OIDC scopes that can be listed in `allowed_scopes`.
var OIDCSupportOptionalScopes = []string{"profile", "email", "address", "phone"}

// ValidateOIDCSupport enforces the cross-field rules the Auth0 Management API applies to an
// `oidc_support` object. Nil arguments mean the attribute is not configured (or not yet known).
// The scopes are passed as plain strings, so the helper does not depend on a specific SDK version.
func ValidateOIDCSupport(path string, isAllowed, allowAllScopes *bool, scopes []string) error {
	hasScopes := len(scopes) > 0

	for _, scope := range scopes {
		if scope == "openid" {
			return fmt.Errorf("`%s.allowed_scopes` must not contain `openid`, it is implied", path)
		}
		if !isOptionalOIDCScope(scope) {
			return fmt.Errorf("`%s.allowed_scopes` contains %q, expected one of %v", path, scope, OIDCSupportOptionalScopes)
		}
	}

	if isAllowed == nil {
		return nil
	}

	if !*isAllowed {
		if hasScopes || (allowAllScopes != nil && *allowAllScopes) {
			return fmt.Errorf("`%s.allowed_scopes` and `%s.allow_all_scopes` cannot be set when `is_allowed` is false", path, path)
		}
		return nil
	}

	switch {
	case allowAllScopes != nil && *allowAllScopes && hasScopes:
		return fmt.Errorf("`%s.allowed_scopes` and `%s.allow_all_scopes` are mutually exclusive", path, path)
	case allowAllScopes != nil && !*allowAllScopes && !hasScopes:
		return fmt.Errorf("`%s.allow_all_scopes` is false and `%s.allowed_scopes` is not set, "+
			"set one of them to a non-empty value", path, path)
	case !hasScopes && allowAllScopes == nil:
		return errors.New("`" + path + "` requires either `allowed_scopes` or `allow_all_scopes` when `is_allowed` is true")
	}

	return nil
}

func isOptionalOIDCScope(scope string) bool {
	return slices.Contains(OIDCSupportOptionalScopes, scope)
}
