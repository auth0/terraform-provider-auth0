package connection

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewClientResource will return a new auth0_connection_client (1:1) resource.
func NewClientResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the connection on which to enable the client.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the client for which the connection is enabled.",
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
		},
		CreateContext: createConnectionClient,
		ReadContext:   readConnectionClient,
		DeleteContext: deleteConnectionClient,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID(internalSchema.SeparatorColon, "connection_id", "client_id"),
		},
		Description: "With this resource, you can enable a single client on a connection.",
	}
}

func createConnectionClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	connectionID := data.Get("connection_id").(string)

	mutex.Lock(connectionID)
	defer mutex.Unlock(connectionID)

	connection, err := api.Connection.Read(connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	clientID := data.Get("client_id").(string)
	enabledClients := append(connection.GetEnabledClients(), clientID)

	if err := api.Connection.Update(
		connectionID,
		&management.Connection{EnabledClients: &enabledClients},
	); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connectionID + ":" + clientID)

	return readConnectionClient(ctx, data, meta)
}

func readConnectionClient(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)
	clientID := data.Get("client_id").(string)

	connection, err := api.Connection.Read(connectionID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	found := false
	for _, enabledClientID := range connection.GetEnabledClients() {
		if enabledClientID == clientID {
			found = true
		}
	}
	if !found {
		data.SetId("")
		return nil
	}

	result := multierror.Append(
		data.Set("name", connection.GetName()),
		data.Set("strategy", connection.GetStrategy()),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteConnectionClient(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	connectionID := data.Get("connection_id").(string)

	mutex.Lock(connectionID)
	defer mutex.Unlock(connectionID)

	connection, err := api.Connection.Read(connectionID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	clientID := data.Get("client_id").(string)
	var enabledClients []string
	for _, enabledClientID := range connection.GetEnabledClients() {
		if enabledClientID == clientID {
			continue
		}
		enabledClients = append(enabledClients, enabledClientID)
	}

	if err := api.Connection.Update(
		connectionID,
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
