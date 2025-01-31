package tenant

//// import (
//// 	"context"
//// 	"fmt"
//// 	"net/url"
//// 
//// 	"github.com/hashicorp/go-multierror"
//// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
//// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//// 
//// 	"github.com/auth0/terraform-provider-auth0/internal/config"
//// 	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
//// )
//// 
//// // NewDataSource will return a new auth0_tenant data source.
//// func NewDataSource() *schema.Resource {
//// 	return &schema.Resource{
//// 		ReadContext: readTenantForDataSource,
//// 		Description: "Use this data source to access information about the tenant this provider is configured to access.",
//// 		Schema:      dataSourceSchema(),
//// 	}
//// }
//// 
//// func dataSourceSchema() map[string]*schema.Schema {
//// 	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
//// 	dataSourceSchema["domain"] = &schema.Schema{
//// 		Type:        schema.TypeString,
//// 		Computed:    true,
//// 		Description: "Your Auth0 domain name.",
//// 	}
//// 	dataSourceSchema["management_api_identifier"] = &schema.Schema{
//// 		Type:     schema.TypeString,
//// 		Computed: true,
//// 		Description: "The identifier value of the built-in Management API resource server, " +
//// 			"which can be used as an audience when configuring client grants.",
//// 	}
//// 
//// 	return dataSourceSchema
//// }
//// 
//// func readTenantForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
//// 	api := meta.(*config.Config).GetAPI()
//// 
//// 	tenant, err := api.Tenant.Read(ctx)
//// 	if err != nil {
//// 		return diag.FromErr(err)
//// 	}
//// 
//// 	data.SetId(id.UniqueId())
//// 
//// 	apiURL, err := url.Parse(api.URI())
//// 	if err != nil {
//// 		return diag.FromErr(fmt.Errorf("unable to determine management API URL: %w", err))
//// 	}
//// 
//// 	result := multierror.Append(
//// 		data.Set("domain", apiURL.Hostname()),
//// 		data.Set("management_api_identifier", apiURL.String()),
//// 		flattenTenant(data, tenant),
//// 	)
//// 
//// 	return diag.FromErr(result.ErrorOrNil())
//// }
