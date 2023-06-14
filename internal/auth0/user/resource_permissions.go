package user

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

// NewPermissionsResource will return a new auth0_connection_client resource.
func NewPermissionsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the user to associate the permission to.",
			},
			"permissions": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of API permissions granted to the user.",
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
		CreateContext: upsertUserPermissions,
		UpdateContext: upsertUserPermissions,
		ReadContext:   readUserPermissions,
		DeleteContext: deleteUserPermissions,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage all of a user's permissions.",
	}
}

func upsertUserPermissions(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !data.HasChange("permissions") {
		return nil
	}

	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)

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
		if err := api.User.RemovePermissions(userID, rmPermissions); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
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
		if err := api.User.AssignPermissions(userID, addPermissions); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}
	}

	data.SetId(userID)

	return readUserPermissions(ctx, data, meta)
}

func readUserPermissions(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	permissions, err := api.User.Permissions(data.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("user_id", data.Id()),
		data.Set("permissions", flattenUserPermissions(permissions)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteUserPermissions(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)

	permissions := data.Get("permissions").(*schema.Set).List()

	var rmPermissions []*management.Permission
	for _, rmPermission := range permissions {
		permission := rmPermission.(map[string]interface{})
		rmPermissions = append(rmPermissions, &management.Permission{
			Name:                     auth0.String(permission["name"].(string)),
			ResourceServerIdentifier: auth0.String(permission["resource_server_identifier"].(string)),
		})
	}

	if err := api.User.RemovePermissions(userID, rmPermissions); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
