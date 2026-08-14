package client

import (
	"context"
	"errors"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	managementv3client "github.com/auth0/go-auth0/v3/management/client"
	"github.com/auth0/go-auth0/v3/management/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"
	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewClientGrantOrganizationsDataSource will return a new auth0_client_grant_organizations data source (EA only).
func NewClientGrantOrganizationsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readClientGrantOrganizationsForDataSource,
		Description: "Data source to retrieve all organizations associated with a specific client grant " +
			"(EA only).",
		Schema: map[string]*schema.Schema{
			"client_grant_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the client grant to retrieve the associated organizations for.",
			},
			"organizations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of organizations associated with the client grant.",
				Elem:        commons.OrganizationSummaryElem(),
			},
		},
	}
}

func readClientGrantOrganizationsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	clientGrantID := data.Get("client_grant_id").(string)

	organizations, err := fetchAllClientGrantOrganizations(ctx, apiv3, clientGrantID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(clientGrantID)

	flattened := make([]interface{}, 0, len(organizations))
	for _, organization := range organizations {
		flattened = append(flattened, commons.FlattenOrganizationSummary(organization))
	}

	return diag.FromErr(data.Set("organizations", flattened))
}

func fetchAllClientGrantOrganizations(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	clientGrantID string,
) ([]*managementv3.Organization, error) {
	var organizations []*managementv3.Organization

	page, err := apiv3.ClientGrants.Organizations.List(ctx, clientGrantID, &managementv3.ListClientGrantOrganizationsRequestParameters{})
	if err != nil {
		return nil, err
	}
	organizations = append(organizations, page.Results...)

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
		organizations = append(organizations, page.Results...)
	}

	return organizations, nil
}
