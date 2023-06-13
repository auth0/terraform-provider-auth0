package role

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
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
			StateContext: internalSchema.ImportResourceGroupID(internalSchema.SeparatorDoubleColon, "role_id", "resource_server_identifier", "permission"),
		},
		Description: "With this resource, you can manage role permissions (1-1).",
	}
}

func createRolePermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)
	permissionName := data.Get("permission").(string)

	if err := api.Role.AssociatePermissions(roleID, []*management.Permission{
		{
			ResourceServerIdentifier: &resourceServerID,
			Name:                     &permissionName,
		},
	}); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(roleID + internalSchema.SeparatorDoubleColon + resourceServerID + internalSchema.SeparatorDoubleColon + permissionName)

	return readRolePermission(ctx, data, meta)
}

func readRolePermission(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)

	existingPermissions, err := api.Role.Permissions(roleID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	for _, p := range existingPermissions.Permissions {
		if p.GetName() == permissionName && p.GetResourceServerIdentifier() == resourceServerID {
			result := multierror.Append(
				data.Set("description", p.GetDescription()),
				data.Set("resource_server_name", p.GetResourceServerName()),
			)

			return diag.FromErr(result.ErrorOrNil())
		}
	}

	data.SetId("")
	return nil
}

func deleteRolePermission(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)

	if err := api.Role.RemovePermissions(
		roleID,
		[]*management.Permission{
			{
				ResourceServerIdentifier: &resourceServerID,
				Name:                     &permissionName,
			},
		},
	); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
