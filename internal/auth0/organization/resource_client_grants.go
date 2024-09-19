package organization

//
// Import (
//	"context"
//
//	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//
//	"github.com/auth0/terraform-provider-auth0/internal/config"
//	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
//	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
//)
//
//// NewClientGrantsResource will return a new auth0_organization_client_grants resource.
// Func NewClientGrantsResource() *schema.Resource {
//	return &schema.Resource{
//		Description:   "With this resource, you can manage all client grants associated with an organization.",
//		CreateContext: createOrganizationClientGrant,
//		ReadContext:   readOrganizationClientGrants,
//		DeleteContext: deleteOrganizationClientGrant,
//		Importer: &schema.ResourceImporter{
//			StateContext: internalSchema.ImportResourceGroupID("organization_id", "connection_id"),
//		},
//		Schema: map[string]*schema.Schema{
//			"organization_id": {
//				Type:        schema.TypeString,
//				Required:    true,
//				Description: "The ID of the organization to associate the client grant.",
//			},
//			"grant_id": {
//				Type:        schema.TypeString,
//				Required:    true,
//				Description: "A Client Grant ID to add to the organization.",
//			},
//		},
//	}
//}
//
// func createOrganizationClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	api := meta.(*config.Config).GetAPI()
//	organizationID := data.Get("organization_id").(string)
//	grantID := data.Get("grant_id").(string)
//
//	if err := api.Organization.AssociateClientGrant(ctx, organizationID, grantID); err != nil {
//		return diag.FromErr(err)
//	}
//
//	return readOrganizationClientGrants(ctx, data, meta)
//}
//
//func readOrganizationClientGrants(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	api := meta.(*config.Config).GetAPI()
//
//	organizationID := data.Get("organization_id").(string)
//
//	clientGrants, err := fetchAllOrganizationClientGrants(ctx, api, organizationID)
//	if err != nil {
//		return diag.FromErr(internalError.HandleAPIError(data, err))
//	}
//
//	grantId := data.Get("grant_id").(string)
//	for _, grant := range clientGrants {
//		if grant.GetID() == grantId {
//			return nil
//		}
//	}
//
//	data.SetId("")
//	return nil
//}
//
//func deleteOrganizationClientGrant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	api := meta.(*config.Config).GetAPI()
//
//	organizationID := data.Get("organization_id").(string)
//	grantID := data.Get("grant_id").(string)
//
//	if err := api.Organization.DeleteConnection(ctx, organizationID, grantID); err != nil {
//		return diag.FromErr(internalError.HandleAPIError(data, err))
//	}
//
//	return nil
//}.
