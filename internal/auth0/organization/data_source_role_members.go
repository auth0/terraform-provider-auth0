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
)

// NewRoleMembersDataSource will return a new auth0_organization_role_members data source.
func NewRoleMembersDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readOrganizationRoleMembersForDataSource,
		Description: "Data source to retrieve the organization members assigned a specific role within the " +
			"context of an organization. Only members with a direct role assignment are returned; for the " +
			"groups assigned to the role, use `auth0_organization_role_groups`. (EA only)",
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the organization.",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the role to retrieve the assigned members for.",
			},
			"members": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The organization members assigned to the role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the user.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the user.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email address of the user.",
						},
						"picture": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL to a picture for the user.",
						},
					},
				},
			},
		},
	}
}

func readOrganizationRoleMembersForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)
	roleID := data.Get("role_id").(string)

	members, err := fetchAllOrganizationRoleMembers(ctx, apiv3, organizationID, roleID)
	if err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, organizationID, roleID)

	return diag.FromErr(data.Set("members", flattenOrganizationRoleMembers(members)))
}

func fetchAllOrganizationRoleMembers(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	organizationID string,
	roleID string,
) ([]*managementv3.RoleMember, error) {
	var members []*managementv3.RoleMember

	page, err := apiv3.Organizations.Roles.Members.List(
		ctx,
		organizationID,
		roleID,
		&managementv3.ListOrganizationRoleMembersRequestParameters{},
	)
	if err != nil {
		return nil, err
	}
	members = append(members, page.Results...)

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
		members = append(members, page.Results...)
	}

	return members, nil
}

func flattenOrganizationRoleMembers(members []*managementv3.RoleMember) []interface{} {
	result := make([]interface{}, 0, len(members))

	for _, member := range members {
		result = append(result, map[string]interface{}{
			"user_id": member.GetUserID(),
			"name":    member.GetName(),
			"email":   member.GetEmail(),
			"picture": member.GetPicture(),
		})
	}

	return result
}
