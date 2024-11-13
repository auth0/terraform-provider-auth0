package selfserviceprofile

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_self_service_profile resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createSelfServiceProfile,
		ReadContext:   readSelfServiceProfile,
		UpdateContext: updateSelfServiceProfile,
		DeleteContext: deleteSelfServiceProfile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can create and manage Self-Service Profile for a tenant.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Description:  "The name of the self-service Profile",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 140),
				Description:  "The description of the self-service Profile",
			},
			"user_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 20,
				Description: "This array stores the mapping information that will be shown to the user during " +
					"the SS-SSO flow. The user will be prompted to map the attributes on their identity provider " +
					"to ensure the specified attributes get passed to Auth0.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
							Description:  "Attribute’s name on Auth0 side",
						},
						"description": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
							Description:  " A human readable description of the attribute.",
						},
						"is_optional": {
							Type:     schema.TypeBool,
							Required: true,
							Description: "Indicates if this attribute is optional or if it has to be provided " +
								"by the customer for the application to function.",
						},
					},
				},
			},
			"branding": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Field can be used to customize the look and feel of the wizard.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logo_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL of logo to display on login page.",
						},
						"colors": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Configuration settings for colors for branding.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary": {
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
										Description: "Primary button background color in hexadecimal.",
									},
								},
							},
						},
					},
				},
			},
			"allowed_strategies": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"oidc", "samlp", "waad", "google-apps", "adfs",
						"okta", "keycloak-samlp", "pingfederate"},
						false),
				},
				Description: "List of IdP strategies that will be shown to users during the Self-Service SSO flow.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the profile was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the profile was updated.",
			},
		},
	}
}

func createSelfServiceProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	ssp := expandSelfServiceProfiles(data)

	if err := api.SelfServiceProfile.Create(ctx, ssp); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(ssp.GetID())

	return readSelfServiceProfile(ctx, data, meta)
}

func readSelfServiceProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	ssp, err := api.SelfServiceProfile.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenSelfServiceProfile(data, ssp))
}

func updateSelfServiceProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	ssp := expandSelfServiceProfiles(data)

	if err := api.SelfServiceProfile.Update(ctx, data.Id(), ssp); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readSelfServiceProfile(ctx, data, meta)
}

func deleteSelfServiceProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.SelfServiceProfile.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
