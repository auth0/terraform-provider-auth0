package userattributeprofile

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_user_attribute_profile resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createUserAttributeProfile,
		ReadContext:   readUserAttributeProfile,
		UpdateContext: updateUserAttributeProfile,
		DeleteContext: deleteUserAttributeProfile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage User Attribute Profiles within Auth0. " +
			"User Attribute Profiles allow you to define how user attributes are mapped between " +
			"different identity providers and Auth0.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the User Attribute Profile.",
			},
			"user_id": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration for mapping the user ID from identity providers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oidc_mapping": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"sub",
							}, false),
							Description: "The OIDC mapping for the user ID.",
						},
						"saml_mapping": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							Description: "The SAML mapping for the user ID.",
							MinItems:    1,
							MaxItems:    3,
						},
						"scim_mapping": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							Description:  "The SCIM mapping for the user ID.",
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"strategy_overrides": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Strategy-specific overrides for user ID mapping.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"strategy": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The strategy name (e.g., 'oidc', 'samlp', 'ad', etc.).",
									},
									"oidc_mapping": {
										Type:        schema.TypeString,
										Computed:    true,
										Optional:    true,
										Description: "OIDC mapping override for this strategy.",
										ValidateFunc: validation.StringInSlice([]string{
											"email",
											"sub",
											"oid",
										}, false),
									},
									"saml_mapping": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Description: "SAML mapping override for this strategy.",
										MinItems:    1,
										MaxItems:    3,
									},
									"scim_mapping": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										Description:  "SCIM mapping override for this strategy.",
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
								},
							},
						},
					},
				},
			},
			"user_attributes": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of user attribute configurations.",
				MinItems:    1,
				MaxItems:    64,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Name of the user attribute.",
							ValidateFunc: validation.StringLenBetween(1, 50),
						},
						"description": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Description of the user attribute.",
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"label": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Display label for the user attribute.",
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"profile_required": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether the attribute is required in the profile.",
						},
						"auth0_mapping": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The Auth0 mapping for the user attribute.",
							ValidateFunc: validation.StringLenBetween(1, 50),
						},
						"oidc_mapping": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "The OIDC mapping configuration for the user attribute.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mapping": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The OIDC mapping field.",
										ValidateFunc: validation.StringLenBetween(1, 50),
									},
									"display_name": {
										Type:         schema.TypeString,
										Optional:     true,
										Description:  "Display name for the OIDC mapping.",
										ValidateFunc: validation.StringLenBetween(1, 50),
									},
								},
							},
						},
						"saml_mapping": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							Description: "SAML mapping override for this strategy.",
							MinItems:    1,
							MaxItems:    3,
						},
						"scim_mapping": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							Description:  "The SCIM mapping for the user attribute.",
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"strategy_overrides": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Description: "Strategy-specific overrides for user attribute mapping.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"strategy": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The strategy name (e.g., 'oidc', 'samlp', 'ad', etc.).",
										ValidateFunc: validation.StringInSlice([]string{
											"ad",
											"adfs",
											"google-apps",
											"oidc",
											"okta",
											"pingfederate",
											"samlp",
											"waad",
										}, false),
									},
									"oidc_mapping": {
										Type:        schema.TypeList,
										Optional:    true,
										Computed:    true,
										MaxItems:    1,
										Description: "OIDC mapping override for this strategy.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mapping": {
													Type:         schema.TypeString,
													Required:     true,
													Description:  "The OIDC mapping field.",
													ValidateFunc: validation.StringLenBetween(1, 50),
												},
												"display_name": {
													Type:         schema.TypeString,
													Optional:     true,
													Description:  "Display name for the OIDC mapping.",
													ValidateFunc: validation.StringLenBetween(1, 50),
												},
											},
										},
									},
									"saml_mapping": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Description: "SAML mapping override for this strategy.",
										MinItems:    1,
										MaxItems:    3,
									},
									"scim_mapping": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										Description:  "SCIM mapping override for this strategy.",
										ValidateFunc: validation.StringLenBetween(1, 128),
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

func createUserAttributeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userAttributeProfile, err := expandUserAttributeProfile(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.UserAttributeProfile.Create(ctx, userAttributeProfile); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(userAttributeProfile.GetID())

	return readUserAttributeProfile(ctx, data, meta)
}

func readUserAttributeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userAttributeProfile, err := api.UserAttributeProfile.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenUserAttributeProfile(data, userAttributeProfile))
}

func updateUserAttributeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	// First read the existing profile to get the complete structure.
	existing, err := api.UserAttributeProfile.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	// Expand the new configuration.
	updated, err := expandUserAttributeProfile(data)
	if err != nil {
		return diag.FromErr(err)
	}

	// Merge with existing structure to preserve API-required fields.
	mergeWithExistingProfile(updated, existing, data)

	if err := api.UserAttributeProfile.Update(ctx, data.Id(), updated); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readUserAttributeProfile(ctx, data, meta)
}

// The expand function now only sets fields that are explicitly configured.
func mergeWithExistingProfile(updated, existing *management.UserAttributeProfile, data *schema.ResourceData) {
	// If user configured user_id, we should use their configuration.
	if _, configured := data.GetOk("user_id"); !configured {
		// User didn't configure user_id, so preserve the API's version.
		updated.UserID = existing.UserID
	} else {
		// If saml_mapping is not configured, clear it from the response.
		if samlValues, ok := data.GetOk("user_id.0.saml_mapping"); !ok || len(samlValues.([]interface{})) == 0 {
			updated.UserID.SAMLMapping = &[]string{}
		}
		// If scim_mapping is not configured, clear it from the response.
		if _, ok := data.GetOk("user_id.0.scim_mapping"); !ok {
			updated.UserID.SCIMMapping = nil
		}
	}
	// If user_id is configured, use the expanded version from the user's config.
}

func deleteUserAttributeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.UserAttributeProfile.Delete(ctx, data.Id()); err != nil {
		if internalError.IsStatusNotFound(err) {
			data.SetId("")
			return nil
		}
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
