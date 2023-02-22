package branding

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewThemeDataSource will return a new auth0_branding_theme data source.
func NewThemeDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readBrandingTheme,
		Description: "Use this data source to access information about the tenant's branding theme settings.",
		Schema:      themeDataSourceSchema(),
	}
}

func themeDataSourceSchema() map[string]*schema.Schema {
	return internalSchema.TransformResourceToDataSource(NewThemeResource().Schema)
}
