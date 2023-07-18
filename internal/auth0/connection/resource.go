package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_connection resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createConnection,
		ReadContext:   readConnection,
		UpdateContext: updateConnection,
		DeleteContext: deleteConnection,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With Auth0, you can define sources of users, otherwise known as connections, " +
			"which may include identity providers (such as Google or LinkedIn), databases, or " +
			"passwordless authentication methods. This resource allows you to configure " +
			"and manage connections to be used with your clients and users.",
		Schema: resourceSchema,
	}
}

func createConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, d, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	if err := api.Connection.Create(ctx, connection); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(connection.GetID())

	return readConnection(ctx, d, m)
}

func readConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	connection, err := api.Connection.Read(ctx, d.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return flattenConnection(d, connection)
}

func updateConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, d, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	if err := api.Connection.Update(ctx, d.Id(), connection); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return readConnection(ctx, d, m)
}

func deleteConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Connection.Delete(ctx, d.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return nil
}
