package user

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

// NewOrganizationsDataSource will return a new auth0_user_organizations data source (EA only).
func NewOrganizationsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readUserOrganizationsForDataSource,
		Description: "Data source to retrieve all organization memberships for a specific Auth0 user " +
			"(EA only).",
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the user to retrieve the organization memberships for.",
			},
			"organizations": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of organizations the user is a member of.",
				Elem:        commons.OrganizationSummaryElem(),
			},
		},
	}
}

func readUserOrganizationsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	userID := data.Get("user_id").(string)

	organizations, err := fetchAllUserOrganizations(ctx, apiv3, userID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(userID)

	flattened := make([]interface{}, 0, len(organizations))
	for _, organization := range organizations {
		flattened = append(flattened, commons.FlattenOrganizationSummary(organization))
	}

	return diag.FromErr(data.Set("organizations", flattened))
}

func fetchAllUserOrganizations(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	userID string,
) ([]*managementv3.Organization, error) {
	var organizations []*managementv3.Organization

	page, err := apiv3.Users.Organizations.List(ctx, userID, &managementv3.ListUserOrganizationsRequestParameters{})
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
