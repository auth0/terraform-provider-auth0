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

// NewMemberRoleResource will return a new auth0_organization_member_role (1:1) resource.
func NewMemberRoleResource() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is used to manage the roles assigned to an organization member.",
		CreateContext: createOrganizationMemberRole,
		ReadContext:   readOrganizationMemberRole,
		DeleteContext: deleteOrganizationMemberRole,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID(internalSchema.SeparatorDoubleColon, "organization_id", "user_id", "role_id"),
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
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(organizationID + internalSchema.SeparatorDoubleColon + userID + internalSchema.SeparatorDoubleColon + roleID)

	return readOrganizationMemberRole(ctx, data, meta)
}

func readOrganizationMemberRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)

	memberRoles, err := api.Organization.MemberRoles(ctx, organizationID, userID)
	if err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	roleID := data.Get("role_id").(string)
	for _, role := range memberRoles.Roles {
		if role.GetID() == roleID {
			result := multierror.Append(
				data.Set("role_name", role.GetName()),
				data.Set("role_description", role.GetDescription()),
			)
			return diag.FromErr(result.ErrorOrNil())
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
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
