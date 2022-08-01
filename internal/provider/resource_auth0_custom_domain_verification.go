package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newCustomDomainVerification() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCustomDomainVerification,
		ReadContext:   readCustomDomainVerification,
		DeleteContext: deleteCustomDomainVerification,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"custom_domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"origin_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cname_api_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func createCustomDomainVerification(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
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

		// The cname_api_key field is only given once: when verification
		// succeeds for the first time. Therefore, we set it on the resource in
		// the creation routine only, and never touch it again.
		if err := d.Set("cname_api_key", customDomainVerification.CNAMEAPIKey); err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return readCustomDomainVerification(ctx, d, m)
}

func readCustomDomainVerification(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	customDomain, err := api.CustomDomain.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("custom_domain_id", customDomain.GetID()),
		d.Set("origin_domain_name", customDomain.OriginDomainName),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteCustomDomainVerification(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
