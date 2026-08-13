package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewClientDataSource will return a new auth0_organization_client data source (EA only).
func NewClientDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readOrganizationClientForDataSource,
		Description: "Data source to retrieve a specific client (application) association for an " +
			"organization, by `organization_id` and `client_id` (EA only).",
		Schema: clientDataSourceSchema(),
	}
}

func clientDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(clientResourceSchema())
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "organization_id", "client_id")

	return dataSourceSchema
}

func readOrganizationClientForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)
	clientID := data.Get("client_id").(string)

	organizationClient, err := apiv3.Organizations.Clients.Get(ctx, organizationID, clientID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	internalSchema.SetResourceGroupID(data, organizationID, clientID)

	return diag.FromErr(flattenOrganizationClient(
		data,
		organizationClient.GetUseForMemberAccess(),
		organizationClient.GetClient(),
	))
}
