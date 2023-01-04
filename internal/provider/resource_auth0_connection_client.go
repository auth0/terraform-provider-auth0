package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	errEmptyConnectionClientID         = fmt.Errorf("ID cannot be empty")
	errInvalidConnectionClientIDFormat = fmt.Errorf("ID must be formated as <connectionID>:<clientID>")
)

func newConnectionClient() *schema.Resource {
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
			StateContext: importConnectionClient,
		},
		Description: "With this resource, you can manage enabled clients on a connection.",
	}
}

func importConnectionClient(
	_ context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	rawID := data.Id()
	if rawID == "" {
		return nil, errEmptyConnectionClientID
	}

	if !strings.Contains(rawID, ":") {
		return nil, errInvalidConnectionClientIDFormat
	}

	idPair := strings.Split(rawID, ":")
	if len(idPair) != 2 {
		return nil, errInvalidConnectionClientIDFormat
	}

	result := multierror.Append(
		data.Set("connection_id", idPair[0]),
		data.Set("client_id", idPair[1]),
	)

	data.SetId(resource.UniqueId())

	return []*schema.ResourceData{data}, result.ErrorOrNil()
}

func createConnectionClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	connectionID := data.Get("connection_id").(string)

	globalMutex.Lock(connectionID)
	defer globalMutex.Unlock(connectionID)

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

	data.SetId(resource.UniqueId())

	return readConnectionClient(ctx, data, meta)
}

func readConnectionClient(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

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
	api := meta.(*management.Management)

	connectionID := data.Get("connection_id").(string)

	globalMutex.Lock(connectionID)
	defer globalMutex.Unlock(connectionID)

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
