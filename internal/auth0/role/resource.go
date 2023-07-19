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

func createRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	role := expandRole(d)

	if err := api.Role.Create(ctx, role); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(role.GetID())

	return readRole(ctx, d, m)
}

func readRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	role, err := api.Role.Read(ctx, d.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return diag.FromErr(flattenRole(d, role))
}

func updateRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	role := expandRole(d)

	if err := api.Role.Update(ctx, d.Id(), role); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return readRole(ctx, d, m)
}

func deleteRole(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Role.Delete(ctx, d.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return nil
}
