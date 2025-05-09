package organization

import (
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

func expandTokenQuota(raw cty.Value) *management.TokenQuota {
	if raw.IsNull() {
		return nil
	}

	var quota *management.TokenQuota

	raw.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		clientCredsValue := config.GetAttr("client_credentials")
		if clientCredsValue.IsNull() {
			return false
		}

		clientCredsValue.ForEachElement(func(_ cty.Value, credsConfig cty.Value) (stop bool) {
			enforce := value.Bool(credsConfig.GetAttr("enforce"))
			perHour := value.Int(credsConfig.GetAttr("per_hour"))
			perDay := value.Int(credsConfig.GetAttr("per_day"))

			quota = &management.TokenQuota{
				ClientCredentials: &management.TokenQuotaClientCredentials{
					Enforce: enforce,
				},
			}

			if perHour != nil && *perHour > 0 {
				quota.ClientCredentials.PerHour = perHour
			}

			if perDay != nil && *perDay > 0 {
				quota.ClientCredentials.PerDay = perDay
			}

			return false
		})

		return false
	})

	return quota
}
