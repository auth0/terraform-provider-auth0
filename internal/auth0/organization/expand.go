package organization

import (
	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandOrganization(data *schema.ResourceData) *management.Organization {
	cfg := data.GetRawConfig()

	organization := &management.Organization{
		Name:        value.String(cfg.GetAttr("name")),
		DisplayName: value.String(cfg.GetAttr("display_name")),
		Branding:    expandOrganizationBranding(cfg.GetAttr("branding")),
		TokenQuota:  expandTokenQuota(cfg.GetAttr("token_quota")),
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

func expandTokenQuota(raw interface{}) *management.TokenQuota {
	if raw == nil {
		return nil
	}

	list := raw.([]interface{})
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	data := list[0].(map[string]interface{})
	clientCreds := data["client_credentials"].([]interface{})
	if len(clientCreds) == 0 || clientCreds[0] == nil {
		return nil
	}

	creds := clientCreds[0].(map[string]interface{})
	quota := &management.TokenQuota{
		ClientCredentials: &management.TokenQuotaClientCredentials{
			Enforce: auth0.Bool(creds["enforce"].(bool)),
		},
	}

	if v, ok := creds["per_hour"].(int); ok && v > 0 {
		quota.ClientCredentials.PerHour = auth0.Int(v)
	}

	if v, ok := creds["per_day"].(int); ok && v > 0 {
		quota.ClientCredentials.PerDay = auth0.Int(v)
	}

	return quota
}
