package customdomain

import (
	"context"

	"github.com/auth0/go-auth0/management"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewCustomDomainsDataSource returns a new auth0_custom_domains data source that allows
// listing custom domains by query filter.
func NewCustomDomainsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readCustomDomainsForDataSource,
		Description: "Data source to retrieve multiple custom domains based on a search query. EA Only.",
		Schema: map[string]*schema.Schema{
			"query": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Search query string to filter custom domains.",
			},
			"custom_domains": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of custom domains matching the search criteria.",
				Elem: &schema.Resource{
					Schema: internalSchema.TransformResourceToDataSource(NewResource().Schema),
				},
			},
		},
	}
}

func readCustomDomainsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	query := data.Get("query").(string)

	var customDomains []*management.CustomDomain
	var from string
	options := []management.RequestOption{
		management.Take(100),
	}

	if query != "" {
		options = append(options, management.Parameter("q", query))
	}

	for {
		if from != "" {
			options = append(options, management.From(from))
		}

		customDomainList, err := api.CustomDomain.ListWithPagination(ctx, options...)
		if err != nil {
			return diag.FromErr(err)
		}

		customDomains = append(customDomains, customDomainList.CustomDomains...)

		if !customDomainList.HasNext() {
			break
		}
		from = customDomainList.Next
	}

	data.SetId("custom-domains")
	if err := flattenCustomDomainList(data, customDomains); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
