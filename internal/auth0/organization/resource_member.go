package organization

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewMemberResource will return a new auth0_organization_member resource.
func NewMemberResource() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is used to manage the assignment of members and their roles within an organization.",
		CreateContext: createOrganizationMember,
		ReadContext:   readOrganizationMember,
		UpdateContext: updateOrganizationMember,
		DeleteContext: deleteOrganizationMember,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourcePairID("organization_id", "user_id"),
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization to assign the member to.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the user to add as an organization member.",
			},
			"roles": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The role ID(s) to assign to the organization member.",
				Optional:    true,
				Deprecated: "Managing roles through this attribute is deprecated and it will be removed in a future version. " +
					"Migrate to the `auth0_organization_member_roles` or the `auth0_organization_member_role` resource to manage organization member roles instead. " +
					"Check the [MIGRATION GUIDE](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) on how to do that.",
			},
		},
	}
}

func createOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	mutex := m.(*config.Config).GetMutex()

	userID := d.Get("user_id").(string)
	orgID := d.Get("organization_id").(string)

	mutex.Lock(orgID)
	if err := api.Organization.AddMembers(orgID, []string{userID}); err != nil {
		return diag.FromErr(err)
	}
	mutex.Unlock(orgID)

	d.SetId(orgID + ":" + userID)

	if err := assignRoles(d, m); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return readOrganizationMember(ctx, d, m)
}

func assignRoles(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("roles") {
		return nil
	}

	userID := d.Get("user_id").(string)
	orgID := d.Get("organization_id").(string)

	toAdd, toRemove := value.Difference(d, "roles")

	if err := addMemberRoles(meta, orgID, userID, toAdd); err != nil {
		return err
	}

	return removeMemberRoles(meta, orgID, userID, toRemove)
}

func removeMemberRoles(meta interface{}, orgID string, userID string, roles []interface{}) error {
	if len(roles) == 0 {
		return nil
	}

	var rolesToRemove []string
	for _, role := range roles {
		rolesToRemove = append(rolesToRemove, role.(string))
	}

	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	mutex.Lock(orgID)
	defer mutex.Unlock(orgID)

	return api.Organization.DeleteMemberRoles(orgID, userID, rolesToRemove)
}

func addMemberRoles(meta interface{}, orgID string, userID string, roles []interface{}) error {
	if len(roles) == 0 {
		return nil
	}

	var rolesToAssign []string
	for _, role := range roles {
		rolesToAssign = append(rolesToAssign, role.(string))
	}

	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	mutex.Lock(orgID)
	defer mutex.Unlock(orgID)

	return api.Organization.AssignMemberRoles(orgID, userID, rolesToAssign)
}

func readOrganizationMember(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	orgID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

	roles, err := api.Organization.MemberRoles(orgID, userID)
	if err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var rolesToSet []string
	for _, role := range roles.Roles {
		rolesToSet = append(rolesToSet, role.GetID())
	}

	err = d.Set("roles", rolesToSet)

	return diag.FromErr(err)
}

func updateOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := assignRoles(d, m); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return readOrganizationMember(ctx, d, m)
}

func deleteOrganizationMember(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	mutex := m.(*config.Config).GetMutex()

	userID := d.Get("user_id").(string)
	orgID := d.Get("organization_id").(string)

	mutex.Lock(orgID)
	defer mutex.Unlock(orgID)

	if err := api.Organization.DeleteMember(orgID, []string{userID}); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
