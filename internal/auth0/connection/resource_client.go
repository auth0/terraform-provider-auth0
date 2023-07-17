package connection

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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
			StateContext: internalSchema.ImportResourceGroupID("connection_id", "client_id"),
		},
		Description: "With this resource, you can enable a single client on a connection.",
	}
}

func createConnectionClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)

	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock(connectionID) // Prevents colliding API requests between other `auth0_connection_client` resource.
	defer mutex.Unlock(connectionID)

	connection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	clientID := data.Get("client_id").(string)
	enabledClients := append(connection.GetEnabledClients(), clientID)
	connectionWithEnabledClients := &management.Connection{EnabledClients: &enabledClients}

	if err := api.Connection.Update(ctx, connectionID, connectionWithEnabledClients); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, connectionID, clientID)

	return readConnectionClient(ctx, data, meta)
}

func readConnectionClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)
	clientID := data.Get("client_id").(string)

	connection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
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

	return diag.FromErr(flattenConnectionClient(data, connection))
}

func deleteConnectionClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)

	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock(connectionID) // Prevents colliding API requests between other `auth0_connection_client` resource.
	defer mutex.Unlock(connectionID)

	connection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	clientID := data.Get("client_id").(string)
	var enabledClients []string
	for _, enabledClientID := range connection.GetEnabledClients() {
		if enabledClientID == clientID {
			continue
		}
		enabledClients = append(enabledClients, enabledClientID)
	}

	connectionWithEnabledClients := &management.Connection{EnabledClients: &enabledClients}

	if err := api.Connection.Update(ctx, connectionID, connectionWithEnabledClients); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
