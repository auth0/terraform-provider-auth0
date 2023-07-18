package branding

import (
	"context"

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
	data.SetId(id.UniqueId())
	return readBranding(ctx, data, meta)
}
