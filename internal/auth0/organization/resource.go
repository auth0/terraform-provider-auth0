package organization

import (
	"context"

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
		},
	}
}

func createOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	organization := expandOrganization(d)

	if err := api.Organization.Create(ctx, organization); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organization.GetID())

	return readOrganization(ctx, d, m)
}

func readOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	organization, err := api.Organization.Read(ctx, d.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return diag.FromErr(flattenOrganization(d, organization))
}

func updateOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	organization := expandOrganization(d)

	if err := api.Organization.Update(ctx, d.Id(), organization); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return readOrganization(ctx, d, m)
}

func deleteOrganization(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Organization.Delete(ctx, d.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return nil
}
