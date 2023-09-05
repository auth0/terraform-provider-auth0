package organization

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewMemberRoleResource will return a new auth0_organization_member_role (1:1) resource.
func NewMemberRoleResource() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is used to manage the roles assigned to an organization member.",
		CreateContext: createOrganizationMemberRole,
		ReadContext:   readOrganizationMemberRole,
		DeleteContext: deleteOrganizationMemberRole,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("organization_id", "user_id", "role_id"),
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user ID of the organization member.",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The role ID to assign to the organization member.",
			},
			"role_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the role.",
			},
			"role_description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the role.",
			},
		},
	}
}

func createOrganizationMemberRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)
	roleID := data.Get("role_id").(string)

	if err := api.Organization.AssignMemberRoles(ctx, organizationID, userID, []string{roleID}); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, organizationID, userID, roleID)

	return readOrganizationMemberRole(ctx, data, meta)
}

func readOrganizationMemberRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)

	var memberRoles []management.OrganizationMemberRole
	var page int
	for {
		memberRoleList, err := api.Organization.MemberRoles(
			ctx,
			organizationID,
			userID,
			management.Page(page),
			management.PerPage(100),
		)
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		memberRoles = append(memberRoles, memberRoleList.Roles...)

		if !memberRoleList.HasNext() {
			break
		}

		page++
	}

	roleID := data.Get("role_id").(string)
	for _, role := range memberRoles {
		if role.GetID() == roleID {
			return diag.FromErr(flattenOrganizationMemberRole(data, role))
		}
	}

	data.SetId("")
	return nil
}

func deleteOrganizationMemberRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)
	roleID := data.Get("role_id").(string)

	if err := api.Organization.DeleteMemberRoles(ctx, organizationID, userID, []string{roleID}); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
