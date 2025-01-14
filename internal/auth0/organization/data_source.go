package organization

import (
	"context"
	"strings"

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
				"is_signup_enabled": {
					Type:     schema.TypeBool,
					Computed: true,
					Description: "Determines whether organization sign-up should be enabled for this " +
						"organization connection. Only applicable for database connections. " +
						"Note: `is_signup_enabled` can only be `true` if `assign_membership_on_login` is `true`.",
				},
				"show_as_button": {
					Type:     schema.TypeBool,
					Computed: true,
					Description: "Determines whether a connection should be displayed on this organization’s " +
						"login prompt. Only applicable for enterprise connections.",
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

	dataSourceSchema["client_grants"] = &schema.Schema{
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Computed:    true,
		Description: "Client Grant ID(s) that are associated to the organization.",
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

	foundClientGrants, err := fetchAllOrganizationClientGrants(ctx, api, foundOrganization.GetID())
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenOrganizationForDataSource(data, foundOrganization, foundConnections, foundMembers, foundClientGrants))
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

func fetchAllOrganizationMembers(
	ctx context.Context,
	api *management.Management,
	organizationID string,
) ([]management.OrganizationMember, error) {
	var foundMembers []management.OrganizationMember
	var from string

	options := []management.RequestOption{
		management.Take(100),
		management.IncludeFields("user_id"),
	}

	for {
		if from != "" {
			options = append(options, management.From(from))
		}

		membersList, err := api.Organization.Members(ctx, organizationID, options...)
		if err != nil {
			return nil, err
		}

		foundMembers = append(foundMembers, membersList.Members...)
		if !membersList.HasNext() {
			break
		}

		from = membersList.Next
	}

	return foundMembers, nil
}

func fetchAllOrganizationClientGrants(
	ctx context.Context,
	api *management.Management,
	organizationID string,
) ([]*management.ClientGrant, error) {
	clientGrantList, err := api.Organization.ClientGrants(ctx, organizationID)
	if err != nil {
		if strings.Contains(err.Error(), "Insufficient scope") {
			return []*management.ClientGrant{}, nil
		}
		return nil, err
	}

	return clientGrantList.ClientGrants, nil
}
