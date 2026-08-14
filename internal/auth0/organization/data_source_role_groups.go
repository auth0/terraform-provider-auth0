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
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewRoleGroupsDataSource will return a new auth0_organization_role_groups data source.
func NewRoleGroupsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readOrganizationRoleGroupsForDataSource,
		Description: "Data source to retrieve the groups assigned to a specific role within the context of " +
			"an organization. For the members with a direct role assignment, use " +
			"`auth0_organization_role_members`. (EA only)",
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the organization.",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the role to retrieve the assigned groups for.",
			},
			"groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The groups assigned to the role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier for the group.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the group.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the group.",
						},
						"external_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The external identifier of the group, often used for SCIM synchronization.",
						},
						"connection_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the connection the group belongs to, if it is a connection group.",
						},
						"organization_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the organization the group belongs to, if it is an organization group.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ISO 8601 timestamp when the group was created.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ISO 8601 timestamp when the group was last updated.",
						},
					},
				},
			},
		},
	}
}

func readOrganizationRoleGroupsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)
	roleID := data.Get("role_id").(string)

	groups, err := fetchAllOrganizationRoleGroups(ctx, apiv3, organizationID, roleID)
	if err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, organizationID, roleID)

	return diag.FromErr(data.Set("groups", flattenOrganizationRoleGroups(groups)))
}

func fetchAllOrganizationRoleGroups(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	organizationID string,
	roleID string,
) ([]*managementv3.RoleGroup, error) {
	var groups []*managementv3.RoleGroup

	page, err := apiv3.Organizations.Roles.Groups.List(
		ctx,
		organizationID,
		roleID,
		&managementv3.ListOrganizationRoleGroupsRequestParameters{},
	)
	if err != nil {
		return nil, err
	}
	groups = append(groups, page.Results...)

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
		groups = append(groups, page.Results...)
	}

	return groups, nil
}

func flattenOrganizationRoleGroups(groups []*managementv3.RoleGroup) []interface{} {
	result := make([]interface{}, 0, len(groups))

	for _, group := range groups {
		result = append(result, map[string]interface{}{
			"id":              group.GetID(),
			"name":            group.GetName(),
			"description":     group.GetDescription(),
			"external_id":     group.GetExternalID(),
			"connection_id":   group.GetConnectionID(),
			"organization_id": group.GetOrganizationID(),
			"created_at":      value.FormatTime(group.GetCreatedAt()),
			"updated_at":      value.FormatTime(group.GetUpdatedAt()),
		})
	}

	return result
}
