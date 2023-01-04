package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

var (
	errEmptyOrganizationMemberID         = fmt.Errorf("ID cannot be empty")
	errInvalidOrganizationMemberIDFormat = fmt.Errorf("ID must be formated as <organizationID>:<userID>")
)

func newOrganizationMember() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is used to manage the assignment of members and their roles within an organization.",
		CreateContext: createOrganizationMember,
		ReadContext:   readOrganizationMember,
		UpdateContext: updateOrganizationMember,
		DeleteContext: deleteOrganizationMember,
		Importer: &schema.ResourceImporter{
			StateContext: importOrganizationMember,
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the organization to assign the member to.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the user to add as an organization member.",
			},
			"roles": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The role ID(s) to assign to the organization member.",
				Optional:    true,
			},
		},
	}
}

func importOrganizationMember(
	_ context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	rawID := data.Id()
	if rawID == "" {
		return nil, errEmptyOrganizationMemberID
	}

	if !strings.Contains(rawID, ":") {
		return nil, errInvalidOrganizationMemberIDFormat
	}

	idPair := strings.Split(rawID, ":")
	if len(idPair) != 2 {
		return nil, errInvalidOrganizationMemberIDFormat
	}

	result := multierror.Append(
		data.Set("organization_id", idPair[0]),
		data.Set("user_id", idPair[1]),
	)

	data.SetId(resource.UniqueId())

	return []*schema.ResourceData{data}, result.ErrorOrNil()
}

func createOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	userID := d.Get("user_id").(string)
	orgID := d.Get("organization_id").(string)

	if err := api.Organization.AddMembers(orgID, []string{userID}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.UniqueId())

	if err := assignRoles(d, api); err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign roles to organization member: %w", err))
	}

	return readOrganizationMember(ctx, d, m)
}

func assignRoles(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("roles") {
		return nil
	}

	orgID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

	toAdd, toRemove := value.Difference(d, "roles")

	if err := addMemberRoles(orgID, userID, toAdd, api); err != nil {
		return err
	}

	return removeMemberRoles(orgID, userID, toRemove, api)
}

func removeMemberRoles(orgID string, userID string, roles []interface{}, api *management.Management) error {
	if len(roles) == 0 {
		return nil
	}

	var rolesToRemove []string
	for _, role := range roles {
		rolesToRemove = append(rolesToRemove, role.(string))
	}

	err := api.Organization.DeleteMemberRoles(orgID, userID, rolesToRemove)
	if err != nil {
		// Ignore 404 errors as the role may have been deleted prior to un-assigning them from the member.
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}
		return err
	}

	return nil
}

func addMemberRoles(orgID string, userID string, roles []interface{}, api *management.Management) error {
	if len(roles) == 0 {
		return nil
	}

	var rolesToAssign []string
	for _, role := range roles {
		rolesToAssign = append(rolesToAssign, role.(string))
	}

	return api.Organization.AssignMemberRoles(orgID, userID, rolesToAssign)
}

func readOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

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
	api := m.(*management.Management)

	if err := assignRoles(d, api); err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign members to organization. %w", err))
	}

	return readOrganizationMember(ctx, d, m)
}

func deleteOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	orgID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

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
