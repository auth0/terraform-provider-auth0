package connection

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
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
	mutex := meta.(*config.Config).GetMutex()

	connectionID := data.Get("connection_id").(string)

	mutex.Lock(connectionID)
	defer mutex.Unlock(connectionID)

	connection, err := api.Connection.Read(
		connectionID,
		management.IncludeFields("enabled_clients", "strategy", "name"),
	)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	// This is never nil because the enabled clients is a required parameter.
	enabledClients := value.Strings(data.GetRawConfig().GetAttr("enabled_clients"))

	if diagnostics := guardAgainstErasingUnwantedEnabledClients(
		connection.GetID(),
		*enabledClients,
		connection.GetEnabledClients(),
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	data.SetId(connection.GetID())

	if err := api.Connection.Update(
		connectionID,
		&management.Connection{EnabledClients: enabledClients},
	); err != nil {
		return diag.FromErr(err)
	}

	return readConnectionClients(ctx, data, meta)
}

func readConnectionClients(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connection, err := api.Connection.Read(
		data.Id(),
		management.IncludeFields("enabled_clients", "strategy", "name"),
	)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
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
	mutex := meta.(*config.Config).GetMutex()

	mutex.Lock(data.Id())
	defer mutex.Unlock(data.Id())

	if err := api.Connection.Update(
		data.Id(),
		&management.Connection{EnabledClients: value.Strings(data.GetRawConfig().GetAttr("enabled_clients"))},
	); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return readConnectionClients(ctx, data, meta)
}

func deleteConnectionClients(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	mutex.Lock(data.Id())
	defer mutex.Unlock(data.Id())

	enabledClients := make([]string, 0)

	if err := api.Connection.Update(
		data.Id(),
		&management.Connection{EnabledClients: &enabledClients},
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
