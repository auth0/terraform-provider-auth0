package auth0

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newCustomDomainVerification() *schema.Resource {
	return &schema.Resource{
		Create: createCustomDomainVerification,
		Read:   readCustomDomainVerification,
		Delete: deleteCustomDomainVerification,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"custom_domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func createCustomDomainVerification(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	return resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		customDomainVerification, err := api.CustomDomain.Verify(d.Get("custom_domain_id").(string))
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if customDomainVerification.GetStatus() != "ready" {
			return resource.RetryableError(
				fmt.Errorf("custom domain has status %q", customDomainVerification.GetStatus()),
			)
		}

		log.Printf("[INFO] Custom domain %s verified", customDomainVerification.GetDomain())

		d.SetId(customDomainVerification.GetID())

		return resource.NonRetryableError(readCustomDomainVerification(d, m))
	})
}

func readCustomDomainVerification(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	customDomain, err := api.CustomDomain.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	return d.Set("custom_domain_id", customDomain.GetID())
}

func deleteCustomDomainVerification(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
