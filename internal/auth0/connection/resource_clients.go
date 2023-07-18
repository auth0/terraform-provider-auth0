package connection

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewClientsResource will return a new auth0_connection_clients (1:many) resource.
func NewClientsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the connection on which to enable the client.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the connection on which to enable the client.",
			},
			"strategy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The strategy of the connection on which to enable the client.",
			},
			"enabled_clients": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "IDs of the clients for which the connection is enabled.",
			},
		},
		CreateContext: createConnectionClients,
		ReadContext:   readConnectionClients,
		UpdateContext: updateConnectionClients,
		DeleteContext: deleteConnectionClients,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage all of the enabled clients on a connection.",
	}
}

func createConnectionClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)

	connection, err := api.Connection.Read(
		ctx,
		connectionID,
		management.IncludeFields("enabled_clients", "strategy", "name"),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	connectionWithEnabledClients := &management.Connection{
		EnabledClients: value.Strings(data.GetRawConfig().GetAttr("enabled_clients")),
	}

	if diagnostics := guardAgainstErasingUnwantedEnabledClients(
		connection.GetID(),
		connectionWithEnabledClients.GetEnabledClients(),
		connection.GetEnabledClients(),
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	data.SetId(connection.GetID())

	if err := api.Connection.Update(ctx, connection.GetID(), connectionWithEnabledClients); err != nil {
		return diag.FromErr(err)
	}

	return readConnectionClients(ctx, data, meta)
}

func readConnectionClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	requestOption := management.IncludeFields("enabled_clients", "strategy", "name")
	connection, err := api.Connection.Read(ctx, data.Id(), requestOption)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	result := multierror.Append(
		data.Set("connection_id", connection.GetID()),
		data.Set("name", connection.GetName()),
		data.Set("strategy", connection.GetStrategy()),
		data.Set("enabled_clients", connection.GetEnabledClients()),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateConnectionClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionWithEnabledClients := &management.Connection{
		EnabledClients: value.Strings(data.GetRawConfig().GetAttr("enabled_clients")),
	}

	if err := api.Connection.Update(ctx, data.Id(), connectionWithEnabledClients); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readConnectionClients(ctx, data, meta)
}

func deleteConnectionClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionWithEnabledClients := &management.Connection{
		EnabledClients: &[]string{},
	}

	if err := api.Connection.Update(ctx, data.Id(), connectionWithEnabledClients); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func guardAgainstErasingUnwantedEnabledClients(
	connectionID string,
	configEnabledClients []string,
	connectionEnabledClients []string,
) diag.Diagnostics {
	if len(connectionEnabledClients) == 0 {
		return nil
	}

	if cmp.Equal(configEnabledClients, connectionEnabledClients) {
		return nil
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Connection with non empty enabled clients",
			Detail: cmp.Diff(configEnabledClients, connectionEnabledClients) +
				fmt.Sprintf("\nThe connection already has enabled clients attached to it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_connection_clients.<given-name> %s'.", connectionID),
		},
	}
}
