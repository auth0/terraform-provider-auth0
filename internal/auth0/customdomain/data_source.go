package customdomain

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_custom_domain data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readCustomDomainForDataSource,
		Description: "Data source to retrieve the custom domain configuration.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	return internalSchema.TransformResourceToDataSource(NewResource().Schema)
}

func readCustomDomainForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customDomains, err := api.CustomDomain.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// At the moment there can only ever
	// be one custom domain configured.
	customDomain := customDomains[0]

	data.SetId(customDomain.GetID())

	result := multierror.Append(
		data.Set("domain", customDomain.GetDomain()),
		data.Set("type", customDomain.GetType()),
		data.Set("primary", customDomain.GetPrimary()),
		data.Set("status", customDomain.GetStatus()),
		data.Set("origin_domain_name", customDomain.GetOriginDomainName()),
		data.Set("custom_client_ip_header", customDomain.GetCustomClientIPHeader()),
		data.Set("tls_policy", customDomain.GetTLSPolicy()),
	)

	if customDomain.Verification != nil {
		result = multierror.Append(result, data.Set("verification", []map[string]interface{}{
			{"methods": customDomain.Verification.Methods},
		}))
	}

	return diag.FromErr(result.ErrorOrNil())
}
