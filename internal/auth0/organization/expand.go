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

func expandOrganizationConnections(cfg cty.Value) []*management.OrganizationConnection {
	connections := make([]*management.OrganizationConnection, 0)

	cfg.ForEachElement(func(_ cty.Value, connectionCfg cty.Value) (stop bool) {
		connections = append(connections, &management.OrganizationConnection{
			ConnectionID:            value.String(connectionCfg.GetAttr("connection_id")),
			AssignMembershipOnLogin: value.Bool(connectionCfg.GetAttr("assign_membership_on_login")),
		})

		return stop
	})

	return connections
}
