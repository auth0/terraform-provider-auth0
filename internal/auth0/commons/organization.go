package commons

import (
	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// OrganizationSummaryElem returns the common read-only nested schema used to represent
// an organization within a list, shared by the `auth0_client_grant_organizations` and
// `auth0_user_organizations` data sources (EA only).
func OrganizationSummaryElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the organization.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the organization.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Friendly name of the organization.",
			},
			"third_party_client_access": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Controls whether this organization can be used in user flows with third-party clients. Available values are `allow` or `block`.",
			},
			"is_app_entitlement_active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this organization's app entitlement is active (EA only).",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Metadata associated with the organization.",
			},
			"branding": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "How to style the login pages for this organization.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logo_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL of logo to display on login page.",
						},
						"colors": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Color scheme used to customize the login pages.",
						},
					},
				},
			},
			"token_quota": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The token quota configuration for this organization.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_credentials": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The token quota configuration for client credentials.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enforce": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether the quota is enforced.",
									},
									"per_day": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of issued tokens per day.",
									},
									"per_hour": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of issued tokens per hour.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// FlattenOrganizationSummary flattens a v3 *managementv3.Organization into the map
// shape produced by OrganizationSummaryElem's schema.
func FlattenOrganizationSummary(organization *managementv3.Organization) map[string]interface{} {
	return map[string]interface{}{
		"organization_id":           organization.GetID(),
		"name":                      organization.GetName(),
		"display_name":              organization.GetDisplayName(),
		"third_party_client_access": string(organization.GetThirdPartyClientAccess()),
		"is_app_entitlement_active": organization.GetIsAppEntitlementActive(),
		"metadata":                  flattenOrganizationMetadataV3(organization.GetMetadata()),
		"branding":                  flattenOrganizationBrandingV3(organization.GetBranding()),
		"token_quota":               flattenTokenQuotaV3(organization.GetTokenQuota()),
	}
}

func flattenOrganizationMetadataV3(metadata managementv3.OrganizationMetadata) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}

	result := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		if value != nil {
			result[key] = *value
		}
	}

	return result
}

func flattenOrganizationBrandingV3(branding managementv3.OrganizationBranding) []interface{} {
	if branding.LogoURL == nil && branding.Colors == nil {
		return nil
	}

	colors := map[string]interface{}{}
	if branding.Colors != nil {
		colors["primary"] = branding.Colors.Primary
		colors["page_background"] = branding.Colors.PageBackground
	}

	return []interface{}{
		map[string]interface{}{
			"logo_url": branding.GetLogoURL(),
			"colors":   colors,
		},
	}
}

func flattenTokenQuotaV3(tokenQuota managementv3.TokenQuota) []interface{} {
	if tokenQuota.ClientCredentials == nil {
		return nil
	}

	clientCredentials := map[string]interface{}{
		"enforce": tokenQuota.ClientCredentials.GetEnforce(),
	}

	if tokenQuota.ClientCredentials.PerHour != nil {
		clientCredentials["per_hour"] = tokenQuota.ClientCredentials.GetPerHour()
	}

	if tokenQuota.ClientCredentials.PerDay != nil {
		clientCredentials["per_day"] = tokenQuota.ClientCredentials.GetPerDay()
	}

	return []interface{}{
		map[string]interface{}{
			"client_credentials": []interface{}{clientCredentials},
		},
	}
}
