package customdomain

import (
	"context"
	"errors"

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
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["custom_domain_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The ID of the Custom Domain.",
		Optional:    true,
	}

	return dataSourceSchema
}

func readCustomDomainForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	customDomainID := data.Get("custom_domain_id").(string)

	if customDomainID != "" {
		data.SetId(customDomainID)

		customDomain, err := api.CustomDomain.Read(ctx, customDomainID)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(flattenCustomDomain(data, customDomain))
	}

	customDomains, err := api.CustomDomain.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	switch len(customDomains) {
	case 0:
		return diag.FromErr(errors.New("no custom domain configured on tenant"))
	case 1:
		customDomain := customDomains[0]
		data.SetId(customDomain.GetID())
		return diag.FromErr(flattenCustomDomain(data, customDomain))
	default:
		return diag.FromErr(errors.New("multiple custom domains found, please specify custom_domain_id"))
	}
}
