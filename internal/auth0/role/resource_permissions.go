package role

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewPermissionsResource will return a new auth0_role_permissions resource.
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
		CreateContext: upsertRolePermissions,
		UpdateContext: upsertRolePermissions,
		ReadContext:   readRolePermissions,
		DeleteContext: deleteRolePermissions,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage role permissions (1-many).",
	}
}

func upsertRolePermissions(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	roleID := data.Get("role_id").(string)

	if !data.HasChange("permissions") {
		return nil
	}

	mutex.Lock(roleID)
	defer mutex.Unlock(roleID)

	toAdd, toRemove := value.Difference(data, "permissions")

	var addPermissions []*management.Permission
	for _, addPermission := range toAdd {
		permission := addPermission.(map[string]interface{})
		addPermissions = append(addPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	if len(addPermissions) > 0 {
		if err := api.Role.AssociatePermissions(roleID, addPermissions); err != nil {
			return diag.FromErr(err)
		}
	}

	var rmPermissions []*management.Permission
	for _, rmPermission := range toRemove {
		permission := rmPermission.(map[string]interface{})
		rmPermissions = append(rmPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	if len(rmPermissions) > 0 {
		if err := api.Role.RemovePermissions(roleID, rmPermissions); err != nil {
			return diag.FromErr(err)
		}
	}

	data.SetId(roleID)

	return readRolePermissions(ctx, data, meta)
}

func readRolePermissions(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)

	permissions, err := api.Role.Permissions(roleID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	err = data.Set("permissions", flattenRolePermissions(permissions.Permissions))

	return diag.FromErr(err)
}

func deleteRolePermissions(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	roleID := data.Get("role_id").(string)

	mutex.Lock(roleID)
	defer mutex.Unlock(roleID)

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

	if err := api.Role.RemovePermissions(
		roleID,
		rmPermissions,
	); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	data.SetId("")
	return nil
}
