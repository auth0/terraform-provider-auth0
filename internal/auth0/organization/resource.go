package organization

import (
	"context"
	"net/http"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_organization resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createOrganization,
		ReadContext:   readOrganization,
		UpdateContext: updateOrganization,
		DeleteContext: deleteOrganization,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "The Organizations feature represents a broad update to the Auth0 platform that allows our " +
			"business-to-business (B2B) customers to better manage their partners and customers, and to " +
			"customize the ways that end-users access their applications. Auth0 customers can use " +
			"Organizations to:\n\n  - Represent their business customers and partners in Auth0 and manage their" +
			"\n    membership.\n  - Configure branded, federated login flows for each business." +
			"\n  - Build administration capabilities into their products, using Organizations" +
			"\n    APIs, so that those businesses can manage their own organizations.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of this organization.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Friendly name of this organization.",
			},
			"branding": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Defines how to style the login pages.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logo_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL of logo to display on login page.",
						},
						"colors": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Color scheme used to customize the login pages.",
						},
					},
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Metadata associated with the organization. Maximum of 10 metadata properties allowed.",
			},
			"token_quota": commons.TokenQuotaSchema(),
		},
	}
}

func createOrganization(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organization := expandOrganization(data)

	if err := api.Organization.Create(ctx, organization); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(organization.GetID())

	return readOrganization(ctx, data, meta)
}

func readOrganization(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organization, err := api.Organization.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenOrganization(data, organization))
}

func updateOrganization(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organization := expandOrganization(data)

	if err := api.Organization.Update(ctx, data.Id(), organization); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	nullFields := fetchNullableFields(data)
	if len(nullFields) != 0 {
		if err := api.Request(ctx, http.MethodPatch, api.URI("organizations", data.Id()), nullFields); err != nil {
			return diag.FromErr(err)
		}
	}

	return readOrganization(ctx, data, meta)
}

func deleteOrganization(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Organization.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
