package organization

import (
	"context"
	"fmt"
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
			StateContext: internalSchema.ImportResourcePairID("organization_id", "user_id"),
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
				Description: "The ID of the user that is an organization member.",
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
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)

	mutex.Lock(organizationID)
	defer mutex.Unlock(organizationID)

	data.SetId(organizationID + ":" + userID)

	if err := assignRoles(data, api); err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign roles to organization member: %w", err))
	}

	return readOrganizationMemberRoles(ctx, data, meta)
}

func readOrganizationMemberRoles(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)
	userID := data.Get("user_id").(string)

	memberRoles, err := api.Organization.MemberRoles(organizationID, userID)
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
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	organizationID := data.Get("organization_id").(string)

	mutex.Lock(organizationID)
	defer mutex.Unlock(organizationID)

	if err := assignRoles(data, api); err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign members to organization: %w", err))
	}

	return readOrganizationMember(ctx, data, meta)
}

func deleteOrganizationMemberRoles(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

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

	mutex.Lock(organizationID)
	defer mutex.Unlock(organizationID)

	if err := api.Organization.DeleteMemberRoles(organizationID, userID, rolesToRemove); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
