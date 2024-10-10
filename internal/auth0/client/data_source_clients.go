package client

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewClientsDataSource will return a new auth0_clients data source.
func NewClientsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readClientsForDataSource,
		Description: "Data source to retrieve a list of Auth0 application clients with optional filtering.",
		Schema: map[string]*schema.Schema{
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter clients by name (partial matches supported).",
			},
			"app_types": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Filter clients by application types.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ValidAppTypes, false),
				},
			},
			"is_first_party": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter clients by first party status.",
			},
			"clients": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of clients matching the filter criteria.",
				Elem: &schema.Resource{
					Schema: coreClientDataSourceSchema(),
				},
			},
		},
	}
}

func coreClientDataSourceSchema() map[string]*schema.Schema {
	clientSchema := dataSourceSchema()

	// Remove unused fields from the client schema.
	fieldsToRemove := []string{
		"client_aliases",
		"logo_uri",
		"oidc_conformant",
		"oidc_backchannel_logout_urls",
		"organization_usage",
		"organization_require_behavior",
		"cross_origin_auth",
		"cross_origin_loc",
		"custom_login_page_on",
		"custom_login_page",
		"form_template",
		"require_pushed_authorization_requests",
		"mobile",
		"initiate_login_uri",
		"native_social_login",
		"refresh_token",
		"signing_keys",
		"encryption_key",
		"sso",
		"sso_disabled",
		"jwt_configuration",
		"addons",
		"default_organization",
		"compliance_level",
		"require_proof_of_possession",
		"token_endpoint_auth_method",
		"signed_request_object",
		"client_authentication_methods",
	}

	for _, field := range fieldsToRemove {
		delete(clientSchema, field)
	}

	return clientSchema
}

func readClientsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	nameFilter := data.Get("name_filter").(string)
	appTypesSet := data.Get("app_types").(*schema.Set)
	isFirstParty := data.Get("is_first_party").(bool)

	appTypes := make([]string, 0, appTypesSet.Len())
	for _, v := range appTypesSet.List() {
		appTypes = append(appTypes, v.(string))
	}

	var clients []*management.Client

	params := []management.RequestOption{
		management.PerPage(100),
	}

	if len(appTypes) > 0 {
		params = append(params, management.Parameter("app_type", strings.Join(appTypes, ",")))
	}
	if isFirstParty {
		params = append(params, management.Parameter("is_first_party", "true"))
	}

	var page int
	for {
		// Add current page parameter.
		params = append(params, management.Page(page))

		list, err := api.Client.List(ctx, params...)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, client := range list.Clients {
			if nameFilter == "" || strings.Contains(client.GetName(), nameFilter) {
				clients = append(clients, client)
			}
		}

		if !list.HasNext() {
			break
		}

		// Remove the page parameter and increment for next iteration.
		params = params[:len(params)-1]
		page++
	}

	filterID := generateFilterID(nameFilter, appTypes, isFirstParty)
	data.SetId(filterID)

	if err := flattenClientList(data, clients); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func generateFilterID(nameFilter string, appTypes []string, isFirstParty bool) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%v-%v", nameFilter, appTypes, isFirstParty)))
	return fmt.Sprintf("clients-%x", h.Sum(nil))
}
