package auth0

import (
	"fmt"
	"net/url"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func newDataConfig() *schema.Resource {
	return &schema.Resource{
		Read: readDataConfig,
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

func readDataConfig(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)

	u, err := url.Parse(api.URI())
	if err != nil {
		return fmt.Errorf("unable to read management API URL from client: %w", err)
	}

	d.SetId(resource.UniqueId())
	d.Set("domain", u.Hostname())
	d.Set("management_api_identifier", u.String())
	return nil
}
