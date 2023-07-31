package role

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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
				Description: "The name of the role.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the role.",
			},
		},
	}
}

func createRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	role := expandRole(data)

	if err := api.Role.Create(ctx, role); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(role.GetID())

	return readRole(ctx, data, meta)
}

func readRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	role, err := api.Role.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenRole(data, role))
}

func updateRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	role := expandRole(data)

	if err := api.Role.Update(ctx, data.Id(), role); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readRole(ctx, data, meta)
}

func deleteRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Role.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
