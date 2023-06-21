package role

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewResource will return a new auth0_role resource.
func NewResource() *schema.Resource {
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
			"`auth0_resource_server`, then associated with roles and optionally, users using this resource.",
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
				Deprecated: "Managing permissions through the `permissions` attribute is deprecated and it will be changed to read-only in a future version. " +
					"Migrate to the `auth0_role_permission` or `auth0_role_permissions` resource to manage role permissions instead. " +
					"Check the [MIGRATION GUIDE](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) for more info.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Name of the permission (scope) configured on the resource server. " +
								"If referencing a scope from an `auth0_resource_server` resource, " +
								"use the `value` property, " +
								"for example `auth0_resource_server.my_resource_server.scopes[0].value`.",
						},
						"resource_server_identifier": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique identifier for the resource server.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the permission.",
						},
						"resource_server_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of resource server that the permission is associated with.",
						},
					},
				},
			},
		},
	}
}

func createRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	role := expandRole(d)
	if err := api.Role.Create(ctx, role); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.GetID())

	d.Partial(true)
	if err := assignRolePermissions(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return readRole(ctx, d, m)
}

func readRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	role, err := api.Role.Read(ctx, d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("name", role.GetName()),
		d.Set("description", role.GetDescription()),
	)

	var permissions []*management.Permission
	var page int
	for {
		permissionList, err := api.Role.Permissions(ctx, d.Id(), management.Page(page))
		if err != nil {
			return diag.FromErr(err)
		}

		permissions = append(permissions, permissionList.Permissions...)

		if !permissionList.HasNext() {
			break
		}

		page++
	}

	result = multierror.Append(
		result,
		d.Set("permissions", flattenRolePermissions(permissions)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	role := expandRole(d)
	if err := api.Role.Update(ctx, d.Id(), role); err != nil {
		return diag.FromErr(err)
	}

	d.Partial(true)
	if err := assignRolePermissions(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return readRole(ctx, d, m)
}

func deleteRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Role.Delete(ctx, d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandRole(d *schema.ResourceData) *management.Role {
	config := d.GetRawConfig()

	return &management.Role{
		Name:        value.String(config.GetAttr("name")),
		Description: value.String(config.GetAttr("description")),
	}
}

func assignRolePermissions(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	toAdd, toRemove := value.Difference(d, "permissions")

	var rmPermissions []*management.Permission
	for _, rmPermission := range toRemove {
		permission := rmPermission.(map[string]interface{})
		rmPermissions = append(rmPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	var addPermissions []*management.Permission
	for _, addPermission := range toAdd {
		permission := addPermission.(map[string]interface{})
		addPermissions = append(addPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	api := m.(*config.Config).GetAPI()

	if len(rmPermissions) > 0 {
		if err := api.Role.RemovePermissions(ctx, d.Id(), rmPermissions); err != nil {
			return err
		}
	}

	if len(addPermissions) > 0 {
		if err := api.Role.AssociatePermissions(ctx, d.Id(), addPermissions); err != nil {
			return err
		}
	}

	return nil
}

func flattenRolePermissions(permissions []*management.Permission) []interface{} {
	var result []interface{}
	for _, permission := range permissions {
		result = append(result, map[string]interface{}{
			"name":                       permission.GetName(),
			"description":                permission.GetDescription(),
			"resource_server_identifier": permission.GetResourceServerIdentifier(),
			"resource_server_name":       permission.GetResourceServerName(),
		})
	}
	return result
}
