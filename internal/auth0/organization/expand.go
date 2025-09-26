package organization

import (
	"github.com/auth0/go-auth0/management"
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

func expandOrganizationConnection(connectionCfg cty.Value) *management.OrganizationConnection {
	return &management.OrganizationConnection{
		ConnectionID:            value.String(connectionCfg.GetAttr("connection_id")),
		AssignMembershipOnLogin: value.Bool(connectionCfg.GetAttr("assign_membership_on_login")),
		IsSignupEnabled:         value.Bool(connectionCfg.GetAttr("is_signup_enabled")),
		ShowAsButton:            value.Bool(connectionCfg.GetAttr("show_as_button")),
	}
}

func expandOrganizationConnections(cfg cty.Value) []*management.OrganizationConnection {
	connections := make([]*management.OrganizationConnection, 0)

	cfg.ForEachElement(func(_ cty.Value, connectionCfg cty.Value) (stop bool) {
		connections = append(connections, expandOrganizationConnection(connectionCfg))

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
		Domain: value.String(cfg.GetAttr("domain")),
		Status: value.String(cfg.GetAttr("status")),
		// Note: ID, VerificationTXT, and VerificationHost are read-only and should not be sent to the API.
	}
}
