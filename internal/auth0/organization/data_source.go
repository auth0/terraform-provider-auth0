package organization

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
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

	return dataSourceSchema
}

func readOrganizationForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	var foundOrganization *management.Organization
	var err error

	organizationID := data.Get("organization_id").(string)
	if organizationID != "" {
		foundOrganization, err = api.Organization.Read(organizationID)
		if err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	} else {
		name := data.Get("name").(string)
		page := 0

	outerLoop:
		for {
			organizations, err := api.Organization.List(management.Page(page), management.PerPage(100))
			if err != nil {
				return diag.FromErr(err)
			}

			for _, organization := range organizations.Organizations {
				if organization.GetName() == name {
					foundOrganization = organization
					break outerLoop
				}
			}

			if !organizations.HasNext() {
				break
			}

			page++
		}

		if foundOrganization == nil {
			return diag.Errorf("No organization found with \"name\" = %q", name)
		}
	}

	data.SetId(foundOrganization.GetID())

	result := multierror.Append(
		data.Set("name", foundOrganization.GetName()),
		data.Set("display_name", foundOrganization.GetDisplayName()),
		data.Set("branding", flattenOrganizationBranding(foundOrganization.GetBranding())),
		data.Set("metadata", foundOrganization.GetMetadata()),
	)

	foundConnections, err := fetchAllOrganizationConnections(api, foundOrganization.GetID())
	if err != nil {
		return diag.FromErr(err)
	}

	result = multierror.Append(
		result,
		data.Set("connections", flattenOrganizationConnections(foundConnections)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func fetchAllOrganizationConnections(api *management.Management, organizationID string) ([]*management.OrganizationConnection, error) {
	var foundConnections []*management.OrganizationConnection
	var page int

	for {
		connections, err := api.Organization.Connections(organizationID, management.Page(page), management.PerPage(100))
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

func flattenOrganizationConnections(connections []*management.OrganizationConnection) []interface{} {
	if connections == nil {
		return nil
	}

	result := make([]interface{}, len(connections))
	for index, connection := range connections {
		result[index] = map[string]interface{}{
			"connection_id":              connection.GetConnectionID(),
			"assign_membership_on_login": connection.GetAssignMembershipOnLogin(),
		}
	}

	return result
}
