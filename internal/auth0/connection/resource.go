package connection

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
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
		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    connectionSchemaV0().CoreConfigSchema().ImpliedType(),
				Upgrade: connectionSchemaUpgradeV0,
				Version: 0,
			},
			{
				Type:    connectionSchemaV1().CoreConfigSchema().ImpliedType(),
				Upgrade: connectionSchemaUpgradeV1,
				Version: 1,
			},
		},
	}
}

func createConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, d, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	if err := api.Connection.Create(ctx, connection); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
		return diagnostics
	}

	d.SetId(connection.GetID())

	diagnostics = append(diagnostics, readConnection(ctx, d, m)...)
	return diagnostics
}

func readConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	connection, err := api.Connection.Read(ctx, d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	connectionOptions, diags := flattenConnectionOptions(d, connection.Options)
	if diags.HasError() {
		return diags
	}

	result := multierror.Append(
		d.Set("name", connection.GetName()),
		d.Set("display_name", connection.GetDisplayName()),
		d.Set("is_domain_connection", connection.GetIsDomainConnection()),
		d.Set("strategy", connection.GetStrategy()),
		d.Set("options", connectionOptions),
		d.Set("realms", connection.GetRealms()),
		d.Set("metadata", connection.GetMetadata()),
		d.Set("enabled_clients", connection.GetEnabledClients()),
	)

	switch connection.GetStrategy() {
	case management.ConnectionStrategyGoogleApps,
		management.ConnectionStrategyOIDC,
		management.ConnectionStrategyAD,
		management.ConnectionStrategyAzureAD,
		management.ConnectionStrategySAML,
		management.ConnectionStrategyADFS:
		result = multierror.Append(result, d.Set("show_as_button", connection.GetShowAsButton()))
	}

	diags = append(diags, diag.FromErr(result.ErrorOrNil())...)
	return diags
}

func updateConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, d, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	if err := api.Connection.Update(ctx, d.Id(), connection); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
		return diagnostics
	}

	diagnostics = append(diagnostics, readConnection(ctx, d, m)...)
	return diagnostics
}

func deleteConnection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Connection.Delete(ctx, d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
