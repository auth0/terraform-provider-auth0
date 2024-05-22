package role

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewPermissionsResource will return a new auth0_role_permissions (1:many) resource.
func NewPermissionsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the role to associate the permission to.",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of API permissions granted to the role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of permission.",
						},
						"resource_server_identifier": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Resource server identifier associated with the permission.",
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
		CreateContext: createRolePermissions,
		UpdateContext: updateRolePermissions,
		ReadContext:   readRolePermissions,
		DeleteContext: deleteRolePermissions,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage role permissions (1-many).",
	}
}

func createRolePermissions(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	toAdd := data.Get("permissions").([]interface{})

	var addPermissions []*management.Permission
	for _, permission := range toAdd {
		p := permission.(map[string]interface{})
		addPermissions = append(addPermissions, &management.Permission{
			Name:                     auth0.String(p["name"].(string)),
			ResourceServerIdentifier: auth0.String(p["resource_server_identifier"].(string)),
		})
	}

	if len(addPermissions) > 0 {
		if err := api.Role.AssociatePermissions(ctx, roleID, addPermissions); err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(roleID)

	return readRolePermissions(ctx, data, meta)
}

func updateRolePermissions(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !data.HasChange("permissions") {
		return nil
	}

	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)

	toAdd, toRemove := value.Difference(data, "permissions")

	var rmPermissions []*management.Permission
	for _, rmPermission := range toRemove {
		permission := rmPermission.(map[string]interface{})
		rmPermissions = append(rmPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	if len(rmPermissions) > 0 {
		if err := api.Role.RemovePermissions(ctx, roleID, rmPermissions); err != nil {
			if !internalError.IsStatusNotFound(err) {
				return diag.FromErr(err)
			}
		}
	}

	var addPermissions []*management.Permission
	for _, addPermission := range toAdd {
		permission := addPermission.(map[string]interface{})
		addPermissions = append(addPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	if len(addPermissions) > 0 {
		if err := api.Role.AssociatePermissions(ctx, roleID, addPermissions); err != nil {
			return diag.FromErr(err)
		}
	}

	return readRolePermissions(ctx, data, meta)
}

func readRolePermissions(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	var permissions []*management.Permission
	var page int
	for {
		permissionList, err := api.Role.Permissions(ctx, data.Id(), management.Page(page), management.PerPage(100))
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		permissions = append(permissions, permissionList.Permissions...)

		if !permissionList.HasNext() {
			break
		}

		page++
	}

	return diag.FromErr(flattenRolePermissions(data, permissions))
}

func deleteRolePermissions(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)

	permissionsToRemove := data.Get("permissions").(*schema.Set).List()

	var rmPermissions []*management.Permission
	for _, p := range permissionsToRemove {
		perm := p.(map[string]interface{})
		role := &management.Permission{
			ResourceServerIdentifier: auth0.String(perm["resource_server_identifier"].(string)),
			Name:                     auth0.String(perm["name"].(string)),
		}
		rmPermissions = append(rmPermissions, role)
	}

	if err := api.Role.RemovePermissions(ctx, roleID, rmPermissions); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
