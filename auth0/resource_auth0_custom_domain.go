package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func newCustomDomain() *schema.Resource {
	return &schema.Resource{
		Create: createCustomDomain,
		Read:   readCustomDomain,
		Delete: deleteCustomDomain,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"auth0_managed_certs",
					"self_managed_certs",
				}, true),
			},
			"primary": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verification_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Deprecated:   "The method is chosen according to the type of the custom domain. CNAME for auth0_managed_certs, TXT for self_managed_certs",
				ValidateFunc: validation.StringInSlice([]string{"txt"}, true),
			},
			"verification": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"methods": {
							Type:     schema.TypeList,
							Elem:     schema.TypeMap,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func createCustomDomain(d *schema.ResourceData, m interface{}) error {
	customDomain := buildCustomDomain(d)
	api := m.(*management.Management)
	if err := api.CustomDomain.Create(customDomain); err != nil {
		return err
	}

	d.SetId(auth0.StringValue(customDomain.ID))

	return readCustomDomain(d, m)
}

func readCustomDomain(d *schema.ResourceData, m interface{}) error {
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

	result := multierror.Append(
		d.Set("domain", customDomain.Domain),
		d.Set("type", customDomain.Type),
		d.Set("primary", customDomain.Primary),
		d.Set("status", customDomain.Status),
	)

	if customDomain.Verification != nil {
		result = multierror.Append(result, d.Set("verification", []map[string]interface{}{
			{"methods": customDomain.Verification.Methods},
		}))
	}

	return result.ErrorOrNil()
}

func deleteCustomDomain(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	if err := api.CustomDomain.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func buildCustomDomain(d *schema.ResourceData) *management.CustomDomain {
	return &management.CustomDomain{
		Domain:             String(d, "domain"),
		Type:               String(d, "type"),
		VerificationMethod: String(d, "verification_method"),
	}
}
