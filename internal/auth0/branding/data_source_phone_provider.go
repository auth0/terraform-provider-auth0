package branding

import (
	"context"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewPhoneProviderDataSource creates a new auth0_phone_provider data source.
func NewPhoneProviderDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPhoneProviderForDataSource,
		Description: "Data source to retrieve a specific Auth0 PhoneProvider by `id`.",
		Schema:      phoneProviderDataSourceSchema(),
	}
}

func phoneProviderDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchemas := internalSchema.TransformResourceToDataSource(NewPhoneProviderResource().Schema)
	dataSourceSchemas["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Phone Provider.",
	}

	return dataSourceSchemas
}

func readPhoneProviderForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	phoneProviderID := data.Get("id").(string)
	data.SetId(phoneProviderID)

	phoneProviderConfig, err := api.Branding.ReadPhoneProvider(ctx, phoneProviderID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenPhoneProvider(data, phoneProviderConfig))
}
