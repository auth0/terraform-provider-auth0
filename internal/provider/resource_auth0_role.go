package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: createRole,
		UpdateContext: updateRole,
		ReadContext:   readRole,
		DeleteContext: deleteRole,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can create and manage collections of permissions that can be " +
			"assigned to users, which are otherwise known as roles. Permissions (scopes) are created on " +
			"auth0_resource_server, then associated with roles and optionally, users using this resource.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name for this role.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the role.",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Configuration settings for permissions (scopes) attached to the role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the permission (scope).",
						},
						"resource_server_identifier": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique identifier for the resource server.",
						},
					},
				},
			},
		},
	}
}

func createRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	role := expandRole(d)
	api := m.(*management.Management)
	if err := api.Role.Create(role); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(role.ID))

	// Enable partial state mode. Sub-resources can potentially cause partial
	// state. Therefore, we must explicitly tell Terraform what is safe to
	// persist and what is not.
	//
	// See: https://www.terraform.io/docs/extend/writing-custom-providers.html
	d.Partial(true)
	if err := assignRolePermissions(d, m); err != nil {
		return diag.FromErr(err)
	}
	// We succeeded, disable partial mode.
	// This causes Terraform to save all fields again.
	d.Partial(false)

	return readRole(ctx, d, m)
}

func readRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	role, err := api.Role.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	d.SetId(role.GetID())

	result := multierror.Append(
		d.Set("name", role.Name),
		d.Set("description", role.Description),
	)

	var permissions []*management.Permission
	var page int
	for {
		permissionList, err := api.Role.Permissions(d.Id(), management.Page(page))
		if err != nil {
			return diag.FromErr(err)
		}

		permissions = append(permissions, permissionList.Permissions...)

		if !permissionList.HasNext() {
			break
		}
		page++
	}

	result = multierror.Append(result, d.Set("permissions", flattenRolePermissions(permissions)))

	return diag.FromErr(result.ErrorOrNil())
}

func updateRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	role := expandRole(d)
	api := m.(*management.Management)
	if err := api.Role.Update(d.Id(), role); err != nil {
		return diag.FromErr(err)
	}

	d.Partial(true)
	if err := assignRolePermissions(d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return readRole(ctx, d, m)
}

func deleteRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Role.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func expandRole(d *schema.ResourceData) *management.Role {
	return &management.Role{
		Name:        String(d, "name"),
		Description: String(d, "description"),
	}
}

func assignRolePermissions(d *schema.ResourceData, m interface{}) error {
	add, rm := Diff(d, "permissions")

	var addPermissions []*management.Permission
	for _, addPermission := range add.List() {
		permission := addPermission.(map[string]interface{})
		addPermissions = append(addPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	var rmPermissions []*management.Permission
	for _, rmPermission := range rm.List() {
		permission := rmPermission.(map[string]interface{})
		rmPermissions = append(rmPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	api := m.(*management.Management)

	if len(rmPermissions) > 0 {
		if err := api.Role.RemovePermissions(d.Id(), rmPermissions); err != nil {
			return err
		}
	}

	if len(addPermissions) > 0 {
		if err := api.Role.AssociatePermissions(d.Id(), addPermissions); err != nil {
			return err
		}
	}

	return nil
}

func flattenRolePermissions(permissions []*management.Permission) []interface{} {
	var result []interface{}
	for _, permission := range permissions {
		result = append(result, map[string]interface{}{
			"name":                       permission.Name,
			"resource_server_identifier": permission.ResourceServerIdentifier,
		})
	}
	return result
}
