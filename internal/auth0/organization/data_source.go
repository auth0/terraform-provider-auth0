package organization

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_organization data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readOrganizationForDataSource,
		Description: "Data source to retrieve a specific Auth0 organization by `organization_id` or `name`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["organization_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the organization. If not provided, `name` must be set.",
		AtLeastOneOf: []string{"organization_id", "name"},
	}

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name")
	dataSourceSchema["name"].Description = "The name of the organization. " +
		"If not provided, `organization_id` must be set. " +
		"For performance, it is advised to use the `organization_id` as a lookup if possible."
	dataSourceSchema["name"].AtLeastOneOf = []string{"organization_id", "name"}

	dataSourceSchema["connections"] = &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"connection_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The ID of the enabled connection on the organization.",
				},
				"assign_membership_on_login": {
					Type:     schema.TypeBool,
					Computed: true,
					Description: "When `true`, all users that log in with this connection will be " +
						"automatically granted membership in the organization. When `false`, users must be " +
						"granted membership in the organization before logging in with this connection.",
				},
			},
		},
	}

	dataSourceSchema["members"] = &schema.Schema{
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Computed:    true,
		Description: "User ID(s) that are members of the organization.",
	}

	return dataSourceSchema
}

func readOrganizationForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	foundOrganization, err := findOrganizationByIDOrName(ctx, data, api)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(foundOrganization.GetID())

	foundConnections, err := fetchAllOrganizationConnections(ctx, api, foundOrganization.GetID())
	if err != nil {
		return diag.FromErr(err)
	}

	foundMembers, err := fetchAllOrganizationMembers(ctx, api, foundOrganization.GetID())
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenOrganizationForDataSource(data, foundOrganization, foundConnections, foundMembers))
}

func findOrganizationByIDOrName(
	ctx context.Context,
	data *schema.ResourceData,
	api *management.Management,
) (*management.Organization, error) {
	organizationID := data.Get("organization_id").(string)
	if organizationID != "" {
		return api.Organization.Read(ctx, organizationID)
	}

	organizationName := data.Get("name").(string)
	return api.Organization.ReadByName(ctx, organizationName)
}

func fetchAllOrganizationConnections(ctx context.Context, api *management.Management, organizationID string) ([]*management.OrganizationConnection, error) {
	var foundConnections []*management.OrganizationConnection
	var page int

	for {
		connections, err := api.Organization.Connections(ctx, organizationID, management.Page(page), management.PerPage(100))
		if err != nil {
			return nil, err
		}

		foundConnections = append(foundConnections, connections.OrganizationConnections...)

		if !connections.HasNext() {
			break
		}

		page++
	}

	return foundConnections, nil
}

func fetchAllOrganizationMembers(ctx context.Context, api *management.Management, organizationID string) ([]string, error) {
	foundMembers := make([]string, 0)
	var from string

	for {
		members, err := api.Organization.Members(ctx, organizationID, management.From(from), management.Take(100), management.IncludeFields("user_id"))
		if err != nil {
			return nil, err
		}

		for _, member := range members.Members {
			foundMembers = append(foundMembers, member.GetUserID())
		}

		if !members.HasNext() {
			break
		}

		from = members.Next
	}

	return foundMembers, nil
}
