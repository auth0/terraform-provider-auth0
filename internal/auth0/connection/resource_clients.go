package connection

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
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
	rawEnabledClients := value.Strings(data.GetRawConfig().GetAttr("enabled_clients"))

	// Fetch existing enabled clients from the new API.
	existingClientsResp, err := api.Connection.ReadEnabledClients(ctx, connectionID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	var existingEnabledClientIDs []string
	for _, c := range existingClientsResp.GetClients() {
		existingEnabledClientIDs = append(existingEnabledClientIDs, c.GetClientID())
	}

	// Safety check: disallow overwriting if existing differs from desired state.
	if diagnostics := guardAgainstErasingUnwantedEnabledClients(connectionID, *rawEnabledClients, existingEnabledClientIDs); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	// Build payload to enable each client.
	var payload []management.ConnectionEnabledClient
	for _, clientID := range *rawEnabledClients {
		payload = append(payload, management.ConnectionEnabledClient{
			ClientID: auth0.String(clientID),
			Status:   auth0.Bool(true),
		})
	}

	if len(payload) != 0 {
		if err := api.Connection.UpdateEnabledClients(ctx, connectionID, payload); err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
	}

	data.SetId(connectionID)

	return readConnectionClients(ctx, data, meta)
}

func readConnectionClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	connectionID := data.Id()

	allClients, err := GetAllEnabledClients(ctx, api, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	requestOption := management.IncludeFields("strategy", "name")
	connection, err := api.Connection.Read(ctx, connectionID, requestOption)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenConnectionClients(data, connection, &management.ConnectionEnabledClientList{
		Clients: &allClients,
	}))
}

func updateConnectionClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	connectionID := data.Id()

	oldRaw, newRaw := data.GetChange("enabled_clients")
	oldSet := oldRaw.(*schema.Set)
	newSet := newRaw.(*schema.Set)

	added := newSet.Difference(oldSet).List()
	removed := oldSet.Difference(newSet).List()

	var payload []management.ConnectionEnabledClient

	for _, clientID := range added {
		payload = append(payload, management.ConnectionEnabledClient{
			ClientID: auth0.String(clientID.(string)),
			Status:   auth0.Bool(true),
		})
	}
	for _, clientID := range removed {
		payload = append(payload, management.ConnectionEnabledClient{
			ClientID: auth0.String(clientID.(string)),
			Status:   auth0.Bool(false),
		})
	}

	if err := api.Connection.UpdateEnabledClients(ctx, connectionID, payload); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readConnectionClients(ctx, data, meta)
}

func deleteConnectionClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	connectionID := data.Id()

	set, ok := data.Get("enabled_clients").(*schema.Set)
	if !ok {
		return diag.Errorf("failed to parse enabled_clients from state")
	}

	var payload []management.ConnectionEnabledClient
	for _, raw := range set.List() {
		clientID, ok := raw.(string)
		if !ok {
			return diag.Errorf("failed to cast clientID to string")
		}
		payload = append(payload, management.ConnectionEnabledClient{
			ClientID: auth0.String(clientID),
			Status:   auth0.Bool(false),
		})
	}

	if len(payload) == 0 {
		return nil
	}

	if err := api.Connection.UpdateEnabledClients(ctx, connectionID, payload); err != nil {
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
