package customdomain

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewDefaultResource will return a new auth0_custom_domain_default resource.
func NewDefaultResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCustomDomainDefault,
		ReadContext:   readCustomDomainDefault,
		UpdateContext: updateCustomDomainDefault,
		DeleteContext: deleteCustomDomainDefault,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can configure the default domain for your Auth0 tenant. " +
			"The default domain is the domain that Auth0 will use for various tenant-level operations. " +
			"This resource manages the default domain configuration via the custom domains API.",
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The custom domain name or canonical domain name to set as the default domain for the tenant.",
			},
		},
	}
}

func createCustomDomainDefault(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customDomainDefault := expandCustomDomainDefault(data)

	if err := api.CustomDomain.UpdateDefault(ctx, customDomainDefault); err != nil {
		return diag.FromErr(err)
	}

	// Use a synthetic ID since this is a singleton resource (one per tenant).
	data.SetId("custom_domain_default")

	return readCustomDomainDefault(ctx, data, meta)
}

func readCustomDomainDefault(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customDomainDefault, err := api.CustomDomain.ReadDefault(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenCustomDomainDefault(data, customDomainDefault))
}

func updateCustomDomainDefault(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customDomainDefault := expandCustomDomainDefault(data)

	if err := api.CustomDomain.UpdateDefault(ctx, customDomainDefault); err != nil {
		return diag.FromErr(err)
	}

	return readCustomDomainDefault(ctx, data, meta)
}

func deleteCustomDomainDefault(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// This resource only manages the default domain setting in Terraform state.
	// Deleting it simply removes it from the state without making any API calls,
	// as there is no API endpoint to "unset" the default domain.
	data.SetId("")

	return nil
}

func expandCustomDomainDefault(data *schema.ResourceData) *management.CustomDomainDefault {
	return &management.CustomDomainDefault{
		Domain: auth0.String(data.Get("domain").(string)),
	}
}

func flattenCustomDomainDefault(data *schema.ResourceData, customDomainDefault *management.CustomDomainDefault) error {
	if customDomainDefault == nil {
		return nil
	}

	return data.Set("domain", customDomainDefault.GetDomain())
}
