package customdomain

import (
	"context"

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

	return diag.FromErr(flattenCustomDomain(data, customDomain))
}
