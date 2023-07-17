package role

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewPermissionResource will return a new auth0_role_permission resource.
func NewPermissionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the role to associate the permission to.",
			},
			"permission": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the permission.",
			},
			"resource_server_identifier": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Identifier of the resource server that the permission is associated with.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the permission.",
			},
			"resource_server_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the resource server that the permission is associated with.",
			},
		},
		CreateContext: createRolePermission,
		ReadContext:   readRolePermission,
		DeleteContext: deleteRolePermission,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("role_id", "resource_server_identifier", "permission"),
		},
		Description: "With this resource, you can manage role permissions (1-1).",
	}
}

func createRolePermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)
	permissionName := data.Get("permission").(string)

	if err := api.Role.AssociatePermissions(ctx, roleID, []*management.Permission{
		{
			ResourceServerIdentifier: &resourceServerID,
			Name:                     &permissionName,
		},
	}); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, roleID, resourceServerID, permissionName)

	return readRolePermission(ctx, data, meta)
}

func readRolePermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)

	existingPermissions, err := api.Role.Permissions(ctx, roleID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	for _, permission := range existingPermissions.Permissions {
		if permission.GetName() == permissionName && permission.GetResourceServerIdentifier() == resourceServerID {
			return diag.FromErr(flattenRolePermission(data, permission))
		}
	}

	data.SetId("")
	return nil
}

func deleteRolePermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)

	if err := api.Role.RemovePermissions(
		ctx,
		roleID,
		[]*management.Permission{
			{
				ResourceServerIdentifier: &resourceServerID,
				Name:                     &permissionName,
			},
		},
	); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
