package branding

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_branding data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readBrandingForDataSource,
		Description: "Use this data source to access information about the tenant's branding settings.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	return internalSchema.TransformResourceToDataSource(NewResource().Schema)
}

func readBrandingForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This resource is not identified by an id in the Auth0 management API.
	data.SetId(id.UniqueId())

	api := meta.(*management.Management)

	branding, err := api.Branding.Read()
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("favicon_url", branding.GetFaviconURL()),
		data.Set("logo_url", branding.GetLogoURL()),
		data.Set("colors", flattenBrandingColors(branding.GetColors())),
		data.Set("font", flattenBrandingFont(branding.GetFont())),
	)

	if err := checkForCustomDomains(api); err == nil {
		brandingUniversalLogin, err := flattenBrandingUniversalLogin(api)
		if err != nil {
			return diag.FromErr(err)
		}

		result = multierror.Append(result, data.Set("universal_login", brandingUniversalLogin))
	}

	return diag.FromErr(result.ErrorOrNil())
}
