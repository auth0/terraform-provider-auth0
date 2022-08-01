package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newDataTenant() *schema.Resource {
	return &schema.Resource{
		ReadContext: readDataTenant,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"management_api_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func readDataTenant(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	u, err := url.Parse(api.URI())
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to determine management API URL: %w", err))
	}

	d.SetId(resource.UniqueId())

	var result *multierror.Error
	result = multierror.Append(result, d.Set("domain", u.Hostname()))
	result = multierror.Append(result, d.Set("management_api_identifier", u.String()))

	return diag.FromErr(result.ErrorOrNil())
}
