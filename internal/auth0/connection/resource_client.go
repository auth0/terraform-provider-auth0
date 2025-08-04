package connection

import (
	"context"

	"github.com/auth0/go-auth0"

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
	clientID := data.Get("client_id").(string)

	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock(connectionID)
	defer mutex.Unlock(connectionID)

	payload := []management.ConnectionEnabledClient{
		{
			ClientID: auth0.String(clientID),
			Status:   auth0.Bool(true),
		},
	}

	if err := api.Connection.UpdateEnabledClients(ctx, connectionID, payload); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}
	internalSchema.SetResourceGroupID(data, connectionID, clientID)
	return readConnectionClient(ctx, data, meta)
}

func readConnectionClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)
	clientID := data.Get("client_id").(string)

	// Implement pagination using the Next token.
	var allClients []management.ConnectionEnabledClient
	var next string

	for {
		var enabledClientsResp *management.ConnectionEnabledClientList
		var err error

		if next == "" {
			enabledClientsResp, err = api.Connection.ReadEnabledClients(ctx, connectionID)
		} else {
			enabledClientsResp, err = api.Connection.ReadEnabledClients(ctx, connectionID, management.From(next))
		}

		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		allClients = append(allClients, enabledClientsResp.GetClients()...)

		if !enabledClientsResp.HasNext() {
			break
		}
		next = enabledClientsResp.Next
	}

	found := false
	for _, c := range allClients {
		if c.GetClientID() == clientID {
			found = true
			break
		}
	}

	if !found {
		// Not found or not enabled.
		data.SetId("")
		return nil
	}

	connection, err := api.Connection.Read(ctx, connectionID, management.IncludeFields("strategy", "name"))
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenConnectionClient(data, connection))
}

func deleteConnectionClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)
	clientID := data.Get("client_id").(string)

	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock(connectionID)
	defer mutex.Unlock(connectionID)

	payload := []management.ConnectionEnabledClient{
		{
			ClientID: auth0.String(clientID),
			Status:   auth0.Bool(false),
		},
	}

	if err := api.Connection.UpdateEnabledClients(ctx, connectionID, payload); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
