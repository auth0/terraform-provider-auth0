package organization

import (
	"context"
	"errors"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	managementv3client "github.com/auth0/go-auth0/v3/management/client"
	"github.com/auth0/go-auth0/v3/management/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewClientsDataSource will return a new auth0_organization_clients data source (EA only).
func NewClientsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readOrganizationClientsForDataSource,
		Description: "Data source to retrieve all client (application) associations for a specific " +
			"organization (EA only).",
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the organization to retrieve the client associations for.",
			},
			"clients": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of clients (applications) associated with the organization.",
				Elem: &schema.Resource{
					Schema: organizationClientListItemSchema(),
				},
			},
		},
	}
}

// organizationClientListItemSchema derives the schema for a single item in the
// "clients" list from clientResourceSchema, so the resource, the single-association
// data source and this list data source all share one field definition each.
func organizationClientListItemSchema() map[string]*schema.Schema {
	itemSchema := internalSchema.TransformResourceToDataSource(clientResourceSchema())
	delete(itemSchema, "organization_id")

	return itemSchema
}

func readOrganizationClientsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)

	organizationClients, err := fetchAllOrganizationClients(ctx, apiv3, organizationID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId(organizationID)

	clients := make([]interface{}, 0, len(organizationClients))
	for _, organizationClient := range organizationClients {
		clients = append(clients, flattenOrganizationClientListItem(organizationClient))
	}

	return diag.FromErr(data.Set("clients", clients))
}

func fetchAllOrganizationClients(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	organizationID string,
) ([]*managementv3.OrganizationClient, error) {
	var clients []*managementv3.OrganizationClient

	page, err := apiv3.Organizations.Clients.List(ctx, organizationID, &managementv3.ListOrganizationClientsRequestParameters{})
	if err != nil {
		return nil, err
	}
	clients = append(clients, page.Results...)

	for !page.RawResponse.Done {
		page, err = page.GetNextPage(ctx)
		if page == nil && err == nil {
			break
		}
		if err != nil {
			if errors.Is(err, core.ErrNoPages) {
				break
			}
			return nil, err
		}
		clients = append(clients, page.Results...)
	}

	return clients, nil
}
