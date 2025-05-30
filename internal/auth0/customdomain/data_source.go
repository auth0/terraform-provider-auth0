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
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["custom_domain_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The ID of the Custom Domain.",
		Required:    true,
		//AtLeastOneOf: []string{"organization_id", "name"},.
	}
	return dataSourceSchema
}

func readCustomDomainForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	customDomainID := data.Get("custom_domain_id").(string)
	customDomain, err := api.CustomDomain.Read(ctx, customDomainID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(customDomain.GetID())

	return diag.FromErr(flattenCustomDomain(data, customDomain))
}
