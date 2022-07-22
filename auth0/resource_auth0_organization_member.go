package auth0

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newOrganizationMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: createOrganizationMember,
		ReadContext:   readOrganizationMember,
		UpdateContext: updateOrganizationMember,
		DeleteContext: deleteOrganizationMember,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the organization to assign member to",
			},
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the user to add as organization member",
			},
			"roles": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Role ID(s) to assign to member",
				Optional:    true,
			},
		},
	}
}

func createOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	userID := String(d, "user_id")
	orgID := String(d, "organization_id")

	api := m.(*management.Management)
	if err := api.Organization.AddMembers(*orgID, []string{*userID}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.UniqueId())

	if err := assignRoles(d, m); err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign roles to organization member. %w", err))
	}

	return readOrganizationMember(ctx, d, m)
}

func assignRoles(d *schema.ResourceData, m interface{}) error {
	orgID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

	add, rm := Diff(d, "roles")

	err := addMemberRoles(orgID, userID, add.List(), m)
	if err != nil {
		return err
	}

	err = removeMemberRoles(orgID, userID, rm.List(), m)
	if err != nil {
		return err
	}

	return nil
}

func removeMemberRoles(orgID string, userID string, roles []interface{}, m interface{}) error {
	api := m.(*management.Management)

	rolesToRemove := []string{}
	for _, r := range roles {
		rolesToRemove = append(rolesToRemove, r.(string))
	}
	if len(rolesToRemove) > 0 {
		err := api.Organization.DeleteMemberRoles(orgID, userID, rolesToRemove)
		if err != nil {
			return err
		}
	}
	return nil
}

func addMemberRoles(orgID string, userID string, roles []interface{}, m interface{}) error {
	api := m.(*management.Management)

	rolesToAssign := []string{}
	for _, r := range roles {
		rolesToAssign = append(rolesToAssign, r.(string))
	}
	if len(rolesToAssign) > 0 {
		err := api.Organization.AssignMemberRoles(orgID, userID, rolesToAssign)
		if err != nil {
			return err
		}
	}
	return nil
}

func readOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	orgID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

	roles, err := api.Organization.MemberRoles(orgID, userID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	rolesToSet := []interface{}{}
	for _, role := range roles.Roles {
		rolesToSet = append(rolesToSet, role.ID)
	}

	result := multierror.Append(
		d.Set("roles", rolesToSet),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := assignRoles(d, m); err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign members to organization. %w", err))
	}

	return readOrganizationMember(ctx, d, m)
}

func deleteOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	orgID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

	if err := api.Organization.DeleteMember(orgID, []string{userID}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
