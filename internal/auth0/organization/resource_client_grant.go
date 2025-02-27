package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewOrganizationClientGrantResource will return a new auth0_organization_client_grant resource.
func NewOrganizationClientGrantResource() *schema.Resource {
	return &schema.Resource{
		Description:   "With this resource, you can manage a client grant associated with an organization.",
		CreateContext: createOrganizationClientGrant,
		ReadContext:   readOrganizationClientGrant,
		DeleteContext: deleteOrganizationClientGrant,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("organization_id", "grant_id"),
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization to associate the client grant.",
			},
			"grant_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A Client Grant ID to add to the organization.",
			},
		},
	}
}

func createOrganizationClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	organizationID := data.Get("organization_id").(string)
	grantID := data.Get("grant_id").(string)

	if err := api.Organization.AssociateClientGrant(ctx, organizationID, grantID); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, organizationID, grantID)

	return readOrganizationClientGrant(ctx, data, meta)
}

func readOrganizationClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	clientGrantList, err := api.Organization.ClientGrants(ctx, organizationID)

	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	grantID := data.Get("grant_id").(string)
	for _, grant := range clientGrantList.ClientGrants {
		if grant.GetID() == grantID {
			return nil
		}
	}

	data.SetId("")
	return nil
}

func deleteOrganizationClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	grantID := data.Get("grant_id").(string)

	if err := api.Organization.RemoveClientGrant(ctx, organizationID, grantID); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
