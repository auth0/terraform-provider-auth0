package auth0

import (
	"fmt"
	"net/url"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newDataTenant() *schema.Resource {
	return &schema.Resource{
		Read: readDataTenant,
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

func readDataTenant(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)

	u, err := url.Parse(api.URI())
	if err != nil {
		return fmt.Errorf("unable to determine management API URL: %w", err)
	}

	d.SetId(resource.UniqueId())

	var result *multierror.Error
	result = multierror.Append(result, d.Set("domain", u.Hostname()))
	result = multierror.Append(result, d.Set("management_api_identifier", u.String()))

	return result.ErrorOrNil()
}
