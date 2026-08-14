package organization

import (
	"context"
	"errors"
	"fmt"

	"github.com/auth0/go-auth0"
	managementv3 "github.com/auth0/go-auth0/v3/management"
	managementv3client "github.com/auth0/go-auth0/v3/management/client"
	"github.com/auth0/go-auth0/v3/management/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"
	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewOrganizationsDataSource will return a new auth0_organizations data source.
func NewOrganizationsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readOrganizationsForDataSource,
		Description: "Data source to retrieve all organizations of the tenant. Optionally set " +
			"`include_client_association_for` to also return each organization's entitlement to a " +
			"given client (application).",
		Schema: map[string]*schema.Schema{
			"include_client_association_for": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The ID of a client (application). When set, every returned organization that " +
					"is associated with this client gains a `client` block describing that association (EA only).",
			},
			"organizations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of organizations of the tenant.",
				Elem:        commons.OrganizationSummaryElemWithClientAssociation(),
			},
		},
	}
}

func readOrganizationsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	clientID := data.Get("include_client_association_for").(string)

	organizations, err := fetchAllOrganizations(ctx, apiv3, clientID)
	if err != nil {
		return diag.FromErr(err)
	}

	// The ID is derived from the filter so that changing `include_client_association_for`
	// produces a different data source instance, matching `auth0_clients`.
	if clientID == "" {
		data.SetId("organizations")
	} else {
		data.SetId(fmt.Sprintf("organizations-%s", clientID))
	}

	flattened := make([]interface{}, 0, len(organizations))
	for _, organization := range organizations {
		flattened = append(flattened, commons.FlattenOrganizationSummaryWithClientAssociation(organization))
	}

	return diag.FromErr(data.Set("organizations", flattened))
}

func fetchAllOrganizations(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	clientID string,
) ([]*managementv3.Organization, error) {
	var organizations []*managementv3.Organization

	params := &managementv3.ListOrganizationsRequestParameters{}
	if clientID != "" {
		params.IncludeClientAssociationFor = auth0.String(clientID)
	}

	page, err := apiv3.Organizations.List(ctx, params)
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
