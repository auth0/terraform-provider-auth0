package organization

import (
	"github.com/auth0/go-auth0/management"
	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandOrganization(data *schema.ResourceData) *management.Organization {
	cfg := data.GetRawConfig()

	organization := &management.Organization{
		Name:        value.String(cfg.GetAttr("name")),
		DisplayName: value.String(cfg.GetAttr("display_name")),
		Branding:    expandOrganizationBranding(cfg.GetAttr("branding")),
		TokenQuota:  commons.ExpandTokenQuota(cfg.GetAttr("token_quota")),
	}

	if data.HasChange("metadata") {
		organization.Metadata = value.MapOfStrings(cfg.GetAttr("metadata"))
	}

	return organization
}

func expandOrganizationBranding(brandingList cty.Value) *management.OrganizationBranding {
	var organizationBranding *management.OrganizationBranding

	brandingList.ForEachElement(func(_ cty.Value, branding cty.Value) (stop bool) {
		organizationBranding = &management.OrganizationBranding{
			LogoURL: value.String(branding.GetAttr("logo_url")),
			Colors:  value.MapOfStrings(branding.GetAttr("colors")),
		}

		return stop
	})

	return organizationBranding
}
func expandOrganizationConnectionCreate(data *schema.ResourceData) *managementv2.CreateOrganizationAllConnectionRequestParameters {
	cfg := data.GetRawConfig()
	isEnabled := data.Get("is_enabled").(bool)
	req := &managementv2.CreateOrganizationAllConnectionRequestParameters{
		ConnectionID:               cfg.GetAttr("connection_id").AsString(),
		AssignMembershipOnLogin:    value.Bool(cfg.GetAttr("assign_membership_on_login")),
		IsSignupEnabled:            value.Bool(cfg.GetAttr("is_signup_enabled")),
		ShowAsButton:               value.Bool(cfg.GetAttr("show_as_button")),
		IsEnabled:                  &isEnabled,
		OrganizationConnectionName: value.String(cfg.GetAttr("organization_connection_name")),
	}

	if orgAccessLevel := value.String(cfg.GetAttr("organization_access_level")); orgAccessLevel != nil {
		level := managementv2.OrganizationAccessLevelEnum(*orgAccessLevel)
		req.OrganizationAccessLevel = &level
	}

	return req
}

func expandOrganizationConnectionUpdate(data *schema.ResourceData) *managementv2.UpdateOrganizationConnectionRequestParameters {
	cfg := data.GetRawConfig()
	isEnabled := data.Get("is_enabled").(bool)
	req := &managementv2.UpdateOrganizationConnectionRequestParameters{
		AssignMembershipOnLogin:    value.Bool(cfg.GetAttr("assign_membership_on_login")),
		IsSignupEnabled:            value.Bool(cfg.GetAttr("is_signup_enabled")),
		ShowAsButton:               value.Bool(cfg.GetAttr("show_as_button")),
		IsEnabled:                  &isEnabled,
		OrganizationConnectionName: value.String(cfg.GetAttr("organization_connection_name")),
	}

	if orgAccessLevel := value.String(cfg.GetAttr("organization_access_level")); orgAccessLevel != nil {
		level := managementv2.OrganizationAccessLevelEnumWithNull(*orgAccessLevel)
		req.OrganizationAccessLevel = &level
	}

	return req
}

func expandOrganizationConnectionCreateFromConfig(connectionCfg cty.Value) *managementv2.CreateOrganizationAllConnectionRequestParameters {
	isEnabled := true
	if !connectionCfg.GetAttr("is_enabled").IsNull() {
		isEnabled = connectionCfg.GetAttr("is_enabled").True()
	}

	req := &managementv2.CreateOrganizationAllConnectionRequestParameters{
		ConnectionID:               connectionCfg.GetAttr("connection_id").AsString(),
		AssignMembershipOnLogin:    value.Bool(connectionCfg.GetAttr("assign_membership_on_login")),
		IsSignupEnabled:            value.Bool(connectionCfg.GetAttr("is_signup_enabled")),
		ShowAsButton:               value.Bool(connectionCfg.GetAttr("show_as_button")),
		IsEnabled:                  &isEnabled,
		OrganizationConnectionName: value.String(connectionCfg.GetAttr("organization_connection_name")),
	}

	if orgAccessLevel := value.String(connectionCfg.GetAttr("organization_access_level")); orgAccessLevel != nil {
		level := managementv2.OrganizationAccessLevelEnum(*orgAccessLevel)
		req.OrganizationAccessLevel = &level
	}

	return req
}

func expandOrganizationConnectionsCreate(cfg cty.Value) []*managementv2.CreateOrganizationAllConnectionRequestParameters {
	connections := make([]*managementv2.CreateOrganizationAllConnectionRequestParameters, 0)

	cfg.ForEachElement(func(_ cty.Value, connectionCfg cty.Value) (stop bool) {
		connections = append(connections, expandOrganizationConnectionCreateFromConfig(connectionCfg))

		return stop
	})

	return connections
}

func fetchNullableFields(data *schema.ResourceData) map[string]interface{} {
	type nullCheckFunc func(*schema.ResourceData) bool

	checks := map[string]nullCheckFunc{
		"token_quota": commons.IsTokenQuotaNull,
	}

	nullableMap := make(map[string]interface{})

	for field, checkFunc := range checks {
		if checkFunc(data) {
			nullableMap[field] = nil
		}
	}

	return nullableMap
}

func expandOrganizationDiscoveryDomain(data *schema.ResourceData) *management.OrganizationDiscoveryDomain {
	cfg := data.GetRawConfig()

	return &management.OrganizationDiscoveryDomain{
		Domain:                      value.String(cfg.GetAttr("domain")),
		Status:                      value.String(cfg.GetAttr("status")),
		UseForOrganizationDiscovery: value.Bool(cfg.GetAttr("use_for_organization_discovery")),
		// Note: ID, VerificationTXT, and VerificationHost are read-only and should not be sent to the API.
	}
}

func expandOrganizationDiscoveryDomainFromConfig(domainCfg cty.Value) *management.OrganizationDiscoveryDomain {
	return &management.OrganizationDiscoveryDomain{
		Domain:                      value.String(domainCfg.GetAttr("domain")),
		Status:                      value.String(domainCfg.GetAttr("status")),
		UseForOrganizationDiscovery: value.Bool(domainCfg.GetAttr("use_for_organization_discovery")),
		// Note: ID, VerificationTXT, and VerificationHost are read-only and should not be sent to the API.
	}
}

func expandOrganizationDiscoveryDomains(cfg cty.Value) []*management.OrganizationDiscoveryDomain {
	domains := make([]*management.OrganizationDiscoveryDomain, 0)

	cfg.ForEachElement(func(_ cty.Value, domainCfg cty.Value) (stop bool) {
		domains = append(domains, expandOrganizationDiscoveryDomainFromConfig(domainCfg))

		return stop
	})

	return domains
}
