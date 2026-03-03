package connection

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewConnectionProfileResource will return a new auth0_connection_profile resource.
func NewConnectionProfileResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createConnectionProfile,
		ReadContext:   readConnectionProfile,
		UpdateContext: updateConnectionProfile,
		DeleteContext: deleteConnectionProfile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manage Auth0 connection profiles. Connection profiles allow you to store configuration templates for connections.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the connection profile.",
			},
			"organization": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Organization associated with the connection profile.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"show_as_button": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Whether to show organization as a button.",
						},
						"assign_membership_on_login": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Whether to assign membership on login.",
						},
					},
				},
			},
			"connection_name_prefix_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "con-{org_id}-",
				Description: "Template for generating connection names from the profile.",
			},
			"enabled_features": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of enabled features for the connection profile.",
			},
			"connection_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Connection configuration for the profile.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"strategy_overrides": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
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

func strategyOverrideSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Description: "Strategy override configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled_features": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Enabled features for the strategy override.",
				},
				"connection_config": {
					Type:        schema.TypeList,
					Optional:    true,
					Computed:    true,
					MaxItems:    1,
					Description: "Connection config for the strategy override.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{},
					},
				},
			},
		},
	}
}

func createConnectionProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV2 := meta.(*config.Config).GetAPIV2()

	profile := expandConnectionProfile(data)

	response, err := apiV2.ConnectionProfiles.Create(ctx, profile)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(response.GetID())

	return readConnectionProfile(ctx, data, meta)
}

func readConnectionProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV2 := meta.(*config.Config).GetAPIV2()

	response, err := apiV2.ConnectionProfiles.Get(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		flattenConnectionProfile(data, response),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateConnectionProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV2 := meta.(*config.Config).GetAPIV2()

	profile := expandConnectionProfileForUpdate(data)

	_, err := apiV2.ConnectionProfiles.Update(ctx, data.Id(), profile)
	if err != nil {
		return diag.FromErr(err)
	}

	return readConnectionProfile(ctx, data, meta)
}

func deleteConnectionProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV2 := meta.(*config.Config).GetAPIV2()

	err := apiV2.ConnectionProfiles.Delete(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
