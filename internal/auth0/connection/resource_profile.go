package connection

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// crossAppAccessResourceAppStatusValues are the only values the Auth0 API's
// `cross_app_access_resource_app.status` enum currently accepts.
var crossAppAccessResourceAppStatusValues = []string{"enabled", "disabled"}

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
		Description:   "Manage Auth0 connection profiles. Connection profiles allow you to store configuration templates for connections.",
		CustomizeDiff: validateConnectionProfileCrossAppAccessResourceApp,
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
			"cross_app_access_resource_app": {
				// Deliberately Optional without Computed: Computed would make Terraform's core
				// diffing carry over the prior state whenever the block is omitted from config,
				// so removing the block would never produce a diff and updateConnectionProfile
				// would never be invoked to send the clearing PATCH.
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Description: "Configures the connection profile as a Cross-App Access (XAA) resource application, " +
					"controlling whether organization admins may enable XAA on their identity providers. Requires the " +
					"`my_orgs_cross_app_access_resource_app` tenant flag to be enabled (EA only). Note: this is distinct " +
					"from, and unrelated to, `cross_app_access_resource_app` on `auth0_connection`, which uses a flat " +
					"`status` string rather than this nested `status` block.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "The Cross App Access resource app status configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_value": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(crossAppAccessResourceAppStatusValues, false),
										Description:  "Default status value for organizations that don't have an explicit override. Either `enabled` or `disabled`.",
									},
									"allowed_values": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(crossAppAccessResourceAppStatusValues, false),
										},
										Description: "Status values organizations are allowed to set. When specified, must contain " +
											"both \"enabled\" and \"disabled\" — the API enforces a minimum of 2 unique values drawn " +
											"from a 2-value enum, so any non-empty list must be the full pair. Omit entirely to leave it unrestricted.",
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

// validateConnectionProfileCrossAppAccessResourceApp pre-empts the Auth0 API's `minItems: 2` and
// `uniqueItems: true` rejection of `cross_app_access_resource_app.status.allowed_values` with a
// clearer Terraform-side error, since the only valid non-empty values for a 2-value enum are the
// full pair. This is UX-only: it never rejects a value the API would otherwise accept.
func validateConnectionProfileCrossAppAccessResourceApp(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	rawAllowedValues, ok := diff.GetOk("cross_app_access_resource_app.0.status.0.allowed_values")
	if !ok {
		return nil
	}

	allowedValues, ok := rawAllowedValues.([]interface{})
	if !ok || len(allowedValues) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(allowedValues))
	for _, rawValue := range allowedValues {
		if str, ok := rawValue.(string); ok {
			seen[str] = true
		}
	}

	if len(allowedValues) != len(crossAppAccessResourceAppStatusValues) || !seen["enabled"] || !seen["disabled"] {
		return fmt.Errorf(
			"cross_app_access_resource_app.0.status.0.allowed_values must contain exactly %v: "+
				"the Auth0 API enforces a minimum of 2 unique values drawn from a 2-value enum, so any "+
				"non-empty list must include both values. Omit allowed_values entirely to leave it unrestricted",
			crossAppAccessResourceAppStatusValues,
		)
	}

	return nil
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
	apiV3 := meta.(*config.Config).GetAPIV3()

	profile := expandConnectionProfile(data)

	response, err := apiV3.ConnectionProfiles.Create(ctx, profile)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(response.GetID())

	return readConnectionProfile(ctx, data, meta)
}

func readConnectionProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV3 := meta.(*config.Config).GetAPIV3()

	response, err := apiV3.ConnectionProfiles.Get(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_connection_profile", data, err)
	}

	result := multierror.Append(
		flattenConnectionProfile(data, response),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateConnectionProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV3 := meta.(*config.Config).GetAPIV3()

	profile := expandConnectionProfileForUpdate(data)

	_, err := apiV3.ConnectionProfiles.Update(ctx, data.Id(), profile)
	if err != nil {
		return diag.FromErr(err)
	}

	return readConnectionProfile(ctx, data, meta)
}

func deleteConnectionProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiV3 := meta.(*config.Config).GetAPIV3()

	err := apiV3.ConnectionProfiles.Delete(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
