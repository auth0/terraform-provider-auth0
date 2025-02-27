package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewSCIMConfigurationDataSource will return a new auth0_connection_scim_configuration data source.
func NewSCIMConfigurationDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readSCIMConfigurationDataSource,
		Description: "Data source to retrieve a SCIM configuration for an Auth0 connection by `connection_id`.",
		Schema:      getSCIMDataSourceSchema(),
	}
}

func getSCIMDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewSCIMConfigurationResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "connection_id")
	dataSourceSchema["user_id_attribute"].Description = "User ID attribute for generation unique of user ids."
	dataSourceSchema["mapping"].Description = "Mapping between Auth0 attributes and SCIM attributes."
	dataSourceSchema["mapping"].Optional = true // This is necessary to make the documentation generate correctly.
	dataSourceSchema["default_mapping"] = &schema.Schema{
		Type:         schema.TypeSet,
		Optional:     true, // This is necessary to make the documentation generate correctly.
		RequiredWith: []string{"user_id_attribute", "mapping"},
		Computed:     true,
		Description:  "Default mapping between Auth0 attributes and SCIM attributes for this connection type.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"auth0": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The field location in the Auth0 schema.",
				},
				"scim": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The field location in the SCIM schema.",
				},
			},
		},
	}

	return dataSourceSchema
}

func readSCIMConfigurationDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)
	scimConfiguration, err := api.Connection.ReadSCIMConfiguration(ctx, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	defaultSCIMConfiguration, err := api.Connection.ReadSCIMDefaultConfiguration(ctx, connectionID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	diags := flattenSCIMConfiguration(data, scimConfiguration)
	if diags.HasError() {
		return diags
	}
	err = data.Set("default_mapping", flattenSCIMMappings(defaultSCIMConfiguration.GetMapping()))
	if err == nil {
		data.SetId(connectionID)
	}

	return diag.FromErr(err)
}
