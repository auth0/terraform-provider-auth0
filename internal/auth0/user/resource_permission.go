package user

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/mutex"
)

// NewPermissionResource will return a new auth0_connection_client resource.
func NewPermissionResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of user to associate permission to.",
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
				Description: "The name of the connection on which to enable the client.",
			},
		},
		CreateContext: createUserPermission,
		ReadContext:   readUserPermission,
		DeleteContext: deleteUserPermission,
		// Importer: &schema.ResourceImporter{
		// 	StateContext: importUserPermission,
		// },
		Description: "With this resource, you can manage user permissions.",
	}
}

func createUserPermission(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	userId := data.Get("user_id").(string)
	resourceServerId := data.Get("resource_server_identifier").(string)
	permissionName := data.Get("permission").(string)

	mutex.Global.Lock(userId)
	defer mutex.Global.Unlock(userId)

	if err := api.User.AssignPermissions(userId, []*management.Permission{
		{
			ResourceServerIdentifier: &resourceServerId,
			Name:                     &permissionName,
		},
	}); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(id.UniqueId())

	return readUserPermission(ctx, data, meta)
}

func readUserPermission(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	userId := data.Get("user_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerId := data.Get("resource_server_identifier").(string)

	existingPermissions, err := api.User.Permissions(userId)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	for _, p := range existingPermissions.Permissions {
		if p.GetName() == permissionName && p.GetResourceServerIdentifier() == resourceServerId {
			return nil
		}
	}

	data.SetId("")
	return nil

}

func deleteUserPermission(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	userId := data.Get("user_id").(string)
	permissionName := data.Get("permission").(string)
	resourceServerId := data.Get("resource_server_identifier").(string)

	mutex.Global.Lock(userId)
	defer mutex.Global.Unlock(userId)

	existingPermissions, err := api.User.Permissions(userId)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	found := false
	for _, p := range existingPermissions.Permissions {
		if p.GetName() == permissionName && p.GetResourceServerIdentifier() == resourceServerId {
			found = true
		}
	}

	if !found {
		data.SetId("")
		return nil
	}

	if err := api.User.RemovePermissions(
		userId,
		[]*management.Permission{
			{
				ResourceServerIdentifier: &resourceServerId,
				Name:                     &permissionName,
			},
		},
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
