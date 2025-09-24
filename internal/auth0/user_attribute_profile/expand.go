package user_attribute_profile

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandUserAttributeProfile(data *schema.ResourceData) (*management.UserAttributeProfile, error) {
	cfg := data.GetRawConfig()

	userAttributeProfile := &management.UserAttributeProfile{
		Name: value.String(cfg.GetAttr("name")),
	}

	// Only include user_id if it's actually configured in Terraform
	userIDAttr := cfg.GetAttr("user_id")
	if !userIDAttr.IsNull() && userIDAttr.LengthInt() > 0 {
		userIDAttr.ForEachElement(func(_ cty.Value, userIDCfg cty.Value) (stop bool) {
			userAttributeProfile.UserID = expandUserAttributeProfileUserID(userIDCfg, data)
			return stop
		})
	}

	// Initialize UserAttributes as an empty map
	userAttributes := make(map[string]*management.UserAttributeProfileUserAttributes)
	userAttributeProfile.UserAttributes = userAttributes

	cfg.GetAttr("user_attributes").ForEachElement(func(_ cty.Value, userAttrCfg cty.Value) (stop bool) {
		if attrName := value.String(userAttrCfg.GetAttr("name")); attrName != nil {
			userAttr := expandUserAttributeProfileUserAttribute(userAttrCfg)
			userAttributeProfile.UserAttributes[*attrName] = userAttr
		}
		return false
	})

	return userAttributeProfile, nil
}

func expandUserAttributeProfileUserID(userIDCfg cty.Value, data *schema.ResourceData) *management.UserAttributeProfileUserID {
	userID := &management.UserAttributeProfileUserID{}

	// Always process all mapping fields - if not configured, they'll be cleared
	// This ensures field removal works correctly

	// OIDC Mapping
	if oidcAttr := userIDCfg.GetAttr("oidc_mapping"); !oidcAttr.IsNull() {
		if oidcValue := value.String(oidcAttr); oidcValue != nil && *oidcValue != "" {
			userID.OIDCMapping = oidcValue
		}
	}

	// SAML Mapping - handle field clearing during updates
	samlAttr := userIDCfg.GetAttr("saml_mapping")
	if samlValues := value.Strings(samlAttr); samlValues != nil && len(*samlValues) > 0 {
		userID.SAMLMapping = samlValues
	} else if data.HasChange("user_id.0.saml_mapping") {
		// Field was changed to empty/removed - explicitly clear it
		userID.SAMLMapping = &[]string{}
	}
	// If not configured and not changed, leave as nil

	// SCIM Mapping - only clear if it's explicitly being removed
	scimAttr := userIDCfg.GetAttr("scim_mapping")
	if scimValue := value.String(scimAttr); scimValue != nil && *scimValue != "" {
		userID.SCIMMapping = scimValue
	}
	// Let merge function handle preservation vs clearing

	userIDCfg.GetAttr("strategy_overrides").ForEachElement(func(_ cty.Value, overrideCfg cty.Value) (stop bool) {
		if strategyName := value.String(overrideCfg.GetAttr("strategy")); strategyName != nil {
			if userID.StrategyOverrides == nil {
				overrides := make(map[string]*management.UserAttributeProfileStrategyOverrides)
				userID.StrategyOverrides = overrides
			}
			override := &management.UserAttributeProfileStrategyOverrides{}

			// Only set mapping fields if they are explicitly configured and not empty
			if oidcAttr := overrideCfg.GetAttr("oidc_mapping"); !oidcAttr.IsNull() {
				if oidcValue := value.String(oidcAttr); oidcValue != nil && *oidcValue != "" {
					override.OIDCMapping = oidcValue
				}
				// If null or empty, leave as nil (API will clear the field)
			}
			if samlAttr := overrideCfg.GetAttr("saml_mapping"); !samlAttr.IsNull() {
				if samlValues := value.Strings(samlAttr); samlValues != nil && len(*samlValues) > 0 {
					override.SAMLMapping = samlValues
				}
				// If null or empty, leave as nil (API will clear the field)
			}
			if scimAttr := overrideCfg.GetAttr("scim_mapping"); !scimAttr.IsNull() {
				if scimValue := value.String(scimAttr); scimValue != nil && *scimValue != "" {
					override.SCIMMapping = scimValue
				}
				// If null or empty, leave as nil (API will clear the field)
			}

			userID.StrategyOverrides[*strategyName] = override
		}
		return false
	})

	return userID
}

func expandUserAttributeProfileUserAttribute(userAttrCfg cty.Value) *management.UserAttributeProfileUserAttributes {
	userAttr := &management.UserAttributeProfileUserAttributes{
		Description:     value.String(userAttrCfg.GetAttr("description")),
		Label:           value.String(userAttrCfg.GetAttr("label")),
		ProfileRequired: value.Bool(userAttrCfg.GetAttr("profile_required")),
		Auth0Mapping:    value.String(userAttrCfg.GetAttr("auth0_mapping")),
	}

	// Only set mapping fields if they are explicitly configured and not empty
	if samlAttr := userAttrCfg.GetAttr("saml_mapping"); !samlAttr.IsNull() {
		if samlValues := value.Strings(samlAttr); samlValues != nil && len(*samlValues) > 0 {
			userAttr.SAMLMapping = samlValues
		}
		// If null or empty, leave as nil (API will clear the field)
	}

	if scimAttr := userAttrCfg.GetAttr("scim_mapping"); !scimAttr.IsNull() {
		if scimValue := value.String(scimAttr); scimValue != nil && *scimValue != "" {
			userAttr.SCIMMapping = scimValue
		}
		// If null or empty, leave as nil (API will clear the field)
	}

	// Handle OIDC mapping (can be complex object)
	userAttrCfg.GetAttr("oidc_mapping").ForEachElement(func(_ cty.Value, oidcCfg cty.Value) (stop bool) {
		userAttr.OIDCMapping = &management.UserAttributeProfileOIDCMapping{
			Mapping:     value.String(oidcCfg.GetAttr("mapping")),
			DisplayName: value.String(oidcCfg.GetAttr("display_name")),
		}
		return stop
	})

	// Handle strategy overrides
	userAttrCfg.GetAttr("strategy_overrides").ForEachElement(func(_ cty.Value, overrideCfg cty.Value) (stop bool) {
		if strategyName := value.String(overrideCfg.GetAttr("strategy")); strategyName != nil {
			if userAttr.StrategyOverrides == nil {
				overrides := make(map[string]*management.UserAttributesStrategyOverride)
				userAttr.StrategyOverrides = overrides
			}

			override := management.UserAttributesStrategyOverride{}

			// Only set mapping fields if they are explicitly configured and not empty
			if samlAttr := overrideCfg.GetAttr("saml_mapping"); !samlAttr.IsNull() {
				if samlValues := value.Strings(samlAttr); samlValues != nil && len(*samlValues) > 0 {
					override.SAMLMapping = samlValues
				}
				// If null or empty, leave as nil (API will clear the field)
			}
			if scimAttr := overrideCfg.GetAttr("scim_mapping"); !scimAttr.IsNull() {
				if scimValue := value.String(scimAttr); scimValue != nil && *scimValue != "" {
					override.SCIMMapping = scimValue
				}
				// If null or empty, leave as nil (API will clear the field)
			}

			// Handle OIDC mapping override (can be complex object)
			overrideCfg.GetAttr("oidc_mapping").ForEachElement(func(_ cty.Value, oidcCfg cty.Value) (stop bool) {
				override.OIDCMapping = &management.UserAttributeProfileOIDCMapping{
					Mapping:     value.String(oidcCfg.GetAttr("mapping")),
					DisplayName: value.String(oidcCfg.GetAttr("display_name")),
				}
				return stop
			})

			userAttr.StrategyOverrides[*strategyName] = &override
		}
		return false
	})

	return userAttr
}
