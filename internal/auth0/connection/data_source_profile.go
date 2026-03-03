package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewConnectionProfileDataSource will return a new auth0_connection_profile data source.
func NewConnectionProfileDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readConnectionProfileDataSource,
		Description: "Retrieve information about an Auth0 connection profile.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the connection profile.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the connection profile.",
			},
			"organization": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Organization associated with the connection profile.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"show_as_button": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Whether to show organization as a button.",
						},
						"assign_membership_on_login": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Whether to assign membership on login.",
						},
					},
				},
			},
			"connection_name_prefix_template": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Template for generating connection names from the profile.",
			},
			"enabled_features": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of enabled features for the connection profile.",
			},
			"connection_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Connection configuration for the profile.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"strategy_overrides": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Strategy overrides for the connection profile.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pingfederate": strategyOverrideSchema(),
						"ad":           strategyOverrideSchema(),
						"adfs":         strategyOverrideSchema(),
						"waad":         strategyOverrideSchema(),
						"google_apps":  strategyOverrideSchema(),
						"okta":         strategyOverrideSchema(),
						"oidc":         strategyOverrideSchema(),
						"samlp":        strategyOverrideSchema(),
					},
				},
			},
		},
	}
}

func readConnectionProfileDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV2 := meta.(*config.Config).GetAPIV2()

	profileID := data.Get("id").(string)

	profile, err := apiV2.ConnectionProfiles.Get(ctx, profileID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(profile.GetID())

	if err := flattenConnectionProfile(data, profile); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
