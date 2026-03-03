package branding

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewPhoneNotificationTemplateDataSource creates a new auth0_branding_phone_notification_template data source.
func NewPhoneNotificationTemplateDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPhoneNotificationTemplateForDataSource,
		Description: "Data source to retrieve a specific Auth0 Phone Notification Template by `template_id`.",
		Schema:      phoneNotificationTemplateDataSourceSchema(),
	}
}

func phoneNotificationTemplateDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchemas := internalSchema.TransformResourceToDataSource(NewPhoneNotificationTemplateResource().Schema)
	dataSourceSchemas["template_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Phone Notification Template.",
	}

	return dataSourceSchemas
}

func readPhoneNotificationTemplateForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	templateID := data.Get("template_id").(string)
	data.SetId(templateID)

	template, err := api.Branding.ReadPhoneNotificationTemplate(ctx, templateID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenPhoneNotificationTemplate(data, template))
}
