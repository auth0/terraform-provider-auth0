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
		Schema:        resourceSchema,
		SchemaVersion: 3,
	}
}

func createConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, data, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	if err := api.Connection.Create(ctx, connection); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connection.GetID())

	return readConnection(ctx, data, meta)
}

func readConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connection, err := api.Connection.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return flattenConnection(data, connection)
}

func updateConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, data, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	if err := api.Connection.Update(ctx, data.Id(), connection); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readConnection(ctx, data, meta)
}

func deleteConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Connection.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
