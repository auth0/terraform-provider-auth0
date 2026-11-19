package validation

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// OIDCSupportOptionalScopes are the optional OIDC scopes that can be listed in `allowed_scopes`.
var OIDCSupportOptionalScopes = []string{"profile", "email", "address", "phone"}

// ValidateOIDCSupport rejects the `oidc_support` combinations that the Auth0 Management API rejects.
// The check is skipped while any value in the block is unknown; Terraform runs it again during apply,
// once the values are known and before the write. When `is_allowed` is false, the scope fields must be
// absent, even as `[]` or `false`. When it is true, an empty `allowed_scopes` and a false
// `allow_all_scopes` count as unset, because they are never sent.
func ValidateOIDCSupport(path string, block cty.Value) error {
	if !value.IsSet(block) || !block.IsWhollyKnown() {
		return nil
	}

	isAllowed := value.Bool(value.Attr(block, "is_allowed"))
	if isAllowed == nil {
		return nil
	}

	allowedScopes := value.Attr(block, "allowed_scopes")
	allowAllScopes := value.Attr(block, "allow_all_scopes")

	if !*isAllowed {
		if value.IsSet(allowedScopes) || value.IsSet(allowAllScopes) {
			return fmt.Errorf("`%s.allowed_scopes` and `%s.allow_all_scopes` cannot be set when `is_allowed` is false", path, path)
		}
		return nil
	}

	allowAll := value.Bool(allowAllScopes)
	if allowAll != nil && *allowAll && len(value.SortedStrings(allowedScopes)) > 0 {
		return fmt.Errorf("`%s.allowed_scopes` and `%s.allow_all_scopes` are mutually exclusive", path, path)
	}

	return nil
}
