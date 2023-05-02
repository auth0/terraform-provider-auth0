package tenant

import (
	"context"
	"fmt"
	"net/url"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_tenant data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readTenantForDataSource,
		Description: "Use this data source to access information about the tenant this provider is configured to access.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["domain"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Your Auth0 domain name.",
	}
	dataSourceSchema["management_api_identifier"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Description: "The identifier value of the built-in Management API resource server, " +
			"which can be used as an audience when configuring client grants.",
	}

	return dataSourceSchema
}

func readTenantForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	u, err := url.Parse(api.URI())
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to determine management API URL: %w", err))
	}

	// This resource is not identified by an id in the Auth0 management API.
	data.SetId(id.UniqueId())

	result := multierror.Append(
		data.Set("domain", u.Hostname()),
		data.Set("management_api_identifier", u.String()),
	)
	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	return readTenant(ctx, data, meta)
}
