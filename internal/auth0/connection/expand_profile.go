package connection

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandConnectionProfile(data *schema.ResourceData) *management.CreateConnectionProfileRequestContent {
	config := data.GetRawConfig()

	profile := &management.CreateConnectionProfileRequestContent{
		Name: *value.String(config.GetAttr("name")),
	}

	// Expand organization.
	config.GetAttr("organization").ForEachElement(func(_ cty.Value, orgCfg cty.Value) (stop bool) {
		org := &management.ConnectionProfileOrganization{}

		if showAsButton := value.String(orgCfg.GetAttr("show_as_button")); showAsButton != nil && *showAsButton != "" {
			v := management.ConnectionProfileOrganizationShowAsButtonEnum(*showAsButton)
			org.SetShowAsButton(&v)
		}

		if assignMembership := value.String(orgCfg.GetAttr("assign_membership_on_login")); assignMembership != nil && *assignMembership != "" {
			v := management.ConnectionProfileOrganizationAssignMembershipOnLoginEnum(*assignMembership)
			org.SetAssignMembershipOnLogin(&v)
		}

		profile.Organization = org
		return stop
	})

	// Expand connection_name_prefix_template.
	profile.ConnectionNamePrefixTemplate = value.String(config.GetAttr("connection_name_prefix_template"))

	// Expand enabled_features.
	if features := value.Strings(config.GetAttr("enabled_features")); features != nil && len(*features) > 0 {
		enabledFeatures := make(management.ConnectionProfileEnabledFeatures, 0)
		for _, f := range *features {
			if f != "" {
				enabledFeatures = append(enabledFeatures, management.EnabledFeaturesEnum(f))
			}
		}
		profile.EnabledFeatures = &enabledFeatures
	}

	// Expand connection_config.
	config.GetAttr("connection_config").ForEachElement(func(_ cty.Value, _ cty.Value) (stop bool) {
		profile.ConnectionConfig = &management.ConnectionProfileConfig{}
		return stop
	})

	// Expand strategy_overrides.
	config.GetAttr("strategy_overrides").ForEachElement(func(_ cty.Value, overridesCfg cty.Value) (stop bool) {
		overrides := &management.ConnectionProfileStrategyOverrides{}

		// Pingfederate.
		overridesCfg.GetAttr("pingfederate").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Pingfederate = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Ad.
		overridesCfg.GetAttr("ad").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Ad = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Adfs.
		overridesCfg.GetAttr("adfs").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Adfs = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Waad.
		overridesCfg.GetAttr("waad").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Waad = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// GoogleApps.
		overridesCfg.GetAttr("google_apps").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.GoogleApps = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Okta.
		overridesCfg.GetAttr("okta").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Okta = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Oidc.
		overridesCfg.GetAttr("oidc").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Oidc = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Samlp.
		overridesCfg.GetAttr("samlp").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Samlp = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		profile.StrategyOverrides = overrides
		return stop
	})

	return profile
}

func expandConnectionProfileForUpdate(data *schema.ResourceData) *management.UpdateConnectionProfileRequestContent {
	config := data.GetRawConfig()

	profile := &management.UpdateConnectionProfileRequestContent{
		Name: value.String(config.GetAttr("name")),
	}

	// Expand organization.
	config.GetAttr("organization").ForEachElement(func(_ cty.Value, orgCfg cty.Value) (stop bool) {
		org := &management.ConnectionProfileOrganization{}

		if showAsButton := value.String(orgCfg.GetAttr("show_as_button")); showAsButton != nil && *showAsButton != "" {
			v := management.ConnectionProfileOrganizationShowAsButtonEnum(*showAsButton)
			org.SetShowAsButton(&v)
		}

		if assignMembership := value.String(orgCfg.GetAttr("assign_membership_on_login")); assignMembership != nil && *assignMembership != "" {
			v := management.ConnectionProfileOrganizationAssignMembershipOnLoginEnum(*assignMembership)
			org.SetAssignMembershipOnLogin(&v)
		}

		profile.Organization = org
		return stop
	})

	// Expand connection_name_prefix_template.
	profile.ConnectionNamePrefixTemplate = value.String(config.GetAttr("connection_name_prefix_template"))

	// Expand enabled_features.
	if features := value.Strings(config.GetAttr("enabled_features")); features != nil && len(*features) > 0 {
		enabledFeatures := make(management.ConnectionProfileEnabledFeatures, 0)
		for _, f := range *features {
			if f != "" {
				enabledFeatures = append(enabledFeatures, management.EnabledFeaturesEnum(f))
			}
		}
		profile.EnabledFeatures = &enabledFeatures
	}

	// Expand connection_config.
	config.GetAttr("connection_config").ForEachElement(func(_ cty.Value, _ cty.Value) (stop bool) {
		profile.ConnectionConfig = &management.ConnectionProfileConfig{}
		return stop
	})

	// Expand strategy_overrides.
	config.GetAttr("strategy_overrides").ForEachElement(func(_ cty.Value, overridesCfg cty.Value) (stop bool) {
		overrides := &management.ConnectionProfileStrategyOverrides{}

		// Pingfederate.
		overridesCfg.GetAttr("pingfederate").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Pingfederate = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Ad.
		overridesCfg.GetAttr("ad").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Ad = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Adfs.
		overridesCfg.GetAttr("adfs").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Adfs = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Waad.
		overridesCfg.GetAttr("waad").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Waad = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// GoogleApps.
		overridesCfg.GetAttr("google_apps").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.GoogleApps = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Okta.
		overridesCfg.GetAttr("okta").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Okta = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Oidc.
		overridesCfg.GetAttr("oidc").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Oidc = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		// Samlp.
		overridesCfg.GetAttr("samlp").ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
			overrides.Samlp = expandConnectionProfileStrategyOverride(elem)
			return stop
		})

		profile.StrategyOverrides = overrides
		return stop
	})

	return profile
}

func expandConnectionProfileStrategyOverride(elem cty.Value) *management.ConnectionProfileStrategyOverride {
	override := &management.ConnectionProfileStrategyOverride{}

	// Expand enabled_features.
	if features := value.Strings(elem.GetAttr("enabled_features")); features != nil && len(*features) > 0 {
		enabledFeatures := make(management.ConnectionProfileStrategyOverridesEnabledFeatures, 0)
		for _, f := range *features {
			if f != "" {
				enabledFeatures = append(enabledFeatures, management.EnabledFeaturesEnum(f))
			}
		}
		if len(enabledFeatures) > 0 {
			override.SetEnabledFeatures(&enabledFeatures)
		}
	}

	// Expand connection_config.
	elem.GetAttr("connection_config").ForEachElement(func(_ cty.Value, _ cty.Value) (stop bool) {
		override.SetConnectionConfig(&management.ConnectionProfileStrategyOverridesConnectionConfig{})
		return stop
	})

	return override
}
