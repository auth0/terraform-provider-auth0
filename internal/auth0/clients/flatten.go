package clients

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenClientList(data *schema.ResourceData, clients []*management.Client) error {
	if clients == nil {
		return data.Set("clients", make([]map[string]interface{}, 0))
	}

	clientList := make([]map[string]interface{}, 0, len(clients))
	for _, client := range clients {
		clientMap := map[string]interface{}{
			"client_id":                           client.GetClientID(),
			"client_secret":                       client.GetClientSecret(),
			"name":                                client.GetName(),
			"description":                         client.GetDescription(),
			"app_type":                            client.GetAppType(),
			"is_first_party":                      client.GetIsFirstParty(),
			"is_token_endpoint_ip_header_trusted": client.GetIsTokenEndpointIPHeaderTrusted(),
			"callbacks":                           client.GetCallbacks(),
			"allowed_logout_urls":                 client.GetAllowedLogoutURLs(),
			"allowed_origins":                     client.GetAllowedOrigins(),
			"allowed_clients":                     client.GetAllowedClients(),
			"grant_types":                         client.GetGrantTypes(),
			"web_origins":                         client.GetWebOrigins(),
			"client_metadata":                     client.GetClientMetadata(),
		}
		clientList = append(clientList, clientMap)
	}

	return data.Set("clients", clientList)
}
