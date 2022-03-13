package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Optional: true,
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
	c := buildCustomDomain(d)
	api := m.(*management.Management)
	if err := api.CustomDomain.Create(c); err != nil {
		return err
	}
	d.SetId(auth0.StringValue(c.ID))
	return readCustomDomain(d, m)
}

func readCustomDomain(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	c, err := api.CustomDomain.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(auth0.StringValue(c.ID))
	d.Set("domain", c.Domain)
	d.Set("type", c.Type)
	d.Set("primary", c.Primary)
	d.Set("status", c.Status)

	if c.Verification != nil {
		d.Set("verification", []map[string]interface{}{
			{"methods": c.Verification.Methods},
		})
	}

	return nil
}

func deleteCustomDomain(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	err := api.CustomDomain.Delete(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}
	return err
}

func buildCustomDomain(d *schema.ResourceData) *management.CustomDomain {
	return &management.CustomDomain{
		Domain:             String(d, "domain"),
		Type:               String(d, "type"),
		VerificationMethod: String(d, "verification_method"),
	}
}
