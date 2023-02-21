package branding

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewThemeDataSource will return a new auth0_branding_theme data source.
func NewThemeDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readBrandingThemeForDataSource,
		Description: "Use this data source to access information about the tenant's branding theme settings.",
		Schema:      themeDataSourceSchema(),
	}
}

func themeDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewThemeResource().Schema)

	dataSourceSchema["branding_theme_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the branding theme.",
	}

	return dataSourceSchema
}

func readBrandingThemeForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	brandingThemeID := data.Get("branding_theme_id").(string)
	data.SetId(brandingThemeID)
	return readBrandingTheme(ctx, data, meta)
}
