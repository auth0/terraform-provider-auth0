package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewMemberRolesResource will return a new auth0_organization_member_roles (1:many) resource.
func NewMemberRolesResource() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is used to manage the roles assigned to an organization member.",
		CreateContext: createOrganizationMemberRoles,
		ReadContext:   readOrganizationMemberRoles,
		UpdateContext: updateOrganizationMemberRoles,
		DeleteContext: deleteOrganizationMemberRoles,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("organization_id", "user_id"),
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
			"roles": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The role ID(s) to assign to the organization member.",
				Required:    true,
			},
		},
	}
}

func createOrganizationMemberRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)

	if err := assignMemberRoles(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, organizationID, userID)

	return readOrganizationMemberRoles(ctx, data, meta)
}

func readOrganizationMemberRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)

	memberRoles, err := api.Organization.MemberRoles(ctx, organizationID, userID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	var rolesToSet []string
	for _, role := range memberRoles.Roles {
		rolesToSet = append(rolesToSet, role.GetID())
	}

	return diag.FromErr(data.Set("roles", rolesToSet))
}

func updateOrganizationMemberRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := assignMemberRoles(ctx, data, meta); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readOrganizationMemberRoles(ctx, data, meta)
}

func deleteOrganizationMemberRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)

	roles := data.Get("roles").(*schema.Set).List()
	if len(roles) == 0 {
		return nil
	}

	rolesToRemove := make([]string, 0)
	for _, role := range roles {
		rolesToRemove = append(rolesToRemove, role.(string))
	}

	if err := api.Organization.DeleteMemberRoles(ctx, organizationID, userID, rolesToRemove); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func assignMemberRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("roles") {
		return nil
	}

	userID := d.Get("user_id").(string)
	organizationID := d.Get("organization_id").(string)

	toAdd, toRemove := value.Difference(d, "roles")

	if err := removeMemberRoles(ctx, meta, organizationID, userID, toRemove); err != nil {
		if !internalError.IsStatusNotFound(err) {
			return err
		}
	}

	return addMemberRoles(ctx, meta, organizationID, userID, toAdd)
}

func removeMemberRoles(ctx context.Context, meta interface{}, organizationID string, userID string, roles []interface{}) error {
	if len(roles) == 0 {
		return nil
	}

	var rolesToRemove []string
	for _, role := range roles {
		rolesToRemove = append(rolesToRemove, role.(string))
	}

	api := meta.(*config.Config).GetAPI()

	return api.Organization.DeleteMemberRoles(ctx, organizationID, userID, rolesToRemove)
}

func addMemberRoles(ctx context.Context, meta interface{}, organizationID string, userID string, roles []interface{}) error {
	if len(roles) == 0 {
		return nil
	}

	var rolesToAssign []string
	for _, role := range roles {
		rolesToAssign = append(rolesToAssign, role.(string))
	}

	api := meta.(*config.Config).GetAPI()

	return api.Organization.AssignMemberRoles(ctx, organizationID, userID, rolesToAssign)
}
