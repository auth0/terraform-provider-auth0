package user

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewPermissionResource will return a new auth0_connection_client resource.
func NewPermissionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the user to associate the permission to.",
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
		CreateContext: createUserPermission,
		ReadContext:   readUserPermission,
		DeleteContext: deleteUserPermission,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("user_id", "resource_server_identifier", "permission"),
		},
		Description: "With this resource, you can manage user permissions.",
	}
}

func createUserPermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)
	permissionName := data.Get("permission").(string)

	if err := api.User.AssignPermissions(ctx, userID, []*management.Permission{
		{
			ResourceServerIdentifier: &resourceServerID,
			Name:                     &permissionName,
		},
	}); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, userID, resourceServerID, permissionName)

	return readUserPermission(ctx, data, meta)
}

func readUserPermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)

	existingPermissions, err := api.User.Permissions(ctx, userID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
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

func deleteUserPermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerID := data.Get("resource_server_identifier").(string)

	if err := api.User.RemovePermissions(
		ctx,
		userID,
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
