package resourceserver

import (
	"context"
	"net/http"
	"net/url"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_resource_server data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readResourceServerForDataSource,
		Description: "Data source to retrieve a specific Auth0 resource server by `resource_server_id` or `identifier`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["resource_server_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the resource server. If not provided, `identifier` must be set.",
		AtLeastOneOf: []string{"resource_server_id", "identifier"},
	}

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "identifier")
	dataSourceSchema["identifier"].Description = "The unique identifier for the resource server. " +
		"If not provided, `resource_server_id` must be set."
	dataSourceSchema["identifier"].AtLeastOneOf = []string{"resource_server_id", "identifier"}

	return dataSourceSchema
}

func readResourceServerForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceServerID := data.Get("resource_server_id").(string)
	if resourceServerID == "" {
		resourceServerID = url.PathEscape(data.Get("identifier").(string))
	}

	api := meta.(*config.Config).GetAPI()
	resourceServer, err := api.ResourceServer.Read(ctx, resourceServerID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Ensuring the ID is the resource server ID and not the identifier,
	// as both can be used to find a resource server with the Read() func.
	data.SetId(resourceServer.GetID())

	result := multierror.Append(
		data.Set("name", resourceServer.GetName()),
		data.Set("identifier", resourceServer.GetIdentifier()),
		data.Set("token_lifetime", resourceServer.GetTokenLifetime()),
		data.Set("allow_offline_access", resourceServer.GetAllowOfflineAccess()),
		data.Set("token_lifetime_for_web", resourceServer.GetTokenLifetimeForWeb()),
		data.Set("signing_alg", resourceServer.GetSigningAlgorithm()),
		data.Set("signing_secret", resourceServer.GetSigningSecret()),
		data.Set(
			"skip_consent_for_verifiable_first_party_clients",
			resourceServer.GetSkipConsentForVerifiableFirstPartyClients(),
		),
		data.Set("verification_location", resourceServer.GetVerificationLocation()),
		data.Set("enforce_policies", resourceServer.GetEnforcePolicies()),
		data.Set("token_dialect", resourceServer.GetTokenDialect()),
		data.Set("scopes", flattenResourceServerScopes(resourceServer.GetScopes())),
	)

	return diag.FromErr(result.ErrorOrNil())
}
