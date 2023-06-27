package organization

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
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
			StateContext: internalSchema.ImportResourceGroupID(internalSchema.SeparatorColon, "organization_id", "user_id"),
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
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(organizationID + ":" + userID)

	return readOrganizationMemberRoles(ctx, data, meta)
}

func readOrganizationMemberRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	var rolesToSet []string
	for _, role := range memberRoles.Roles {
		rolesToSet = append(rolesToSet, role.GetID())
	}

	return diag.FromErr(data.Set("roles", rolesToSet))
}

func updateOrganizationMemberRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := assignMemberRoles(ctx, data, meta); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return readOrganizationMember(ctx, data, meta)
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
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
