package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func newCustomDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCustomDomain,
		ReadContext:   readCustomDomain,
		DeleteContext: deleteCustomDomain,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With Auth0, you can use a custom domain to maintain a consistent user experience. " +
			"This resource allows you to create and manage a custom domain within your Auth0 tenant.",
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the custom domain.",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"auth0_managed_certs",
					"self_managed_certs",
				}, true),
				Description: "Provisioning type for the custom domain. " +
					"Options include `auth0_managed_certs` and `self_managed_certs`.",
			},
			"primary": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether this is a primary domain.",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Configuration status for the custom domain. " +
					"Options include `disabled`, `pending`, `pending_verification`, and `ready`.",
			},
			"origin_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Once the configuration status is `ready`, the DNS name " +
					"of the Auth0 origin server that handles traffic for the custom domain.",
			},
			"verification": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Configuration settings for verification.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"methods": {
							Type:        schema.TypeList,
							Elem:        schema.TypeMap,
							Computed:    true,
							Description: "Verification methods for the domain.",
						},
					},
				},
			},
		},
	}
}

func createCustomDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	customDomain := expandCustomDomain(d.GetRawConfig())
	if err := api.CustomDomain.Create(customDomain); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(customDomain.GetID())

	return readCustomDomain(ctx, d, m)
}

func readCustomDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	customDomain, err := api.CustomDomain.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("domain", customDomain.GetDomain()),
		d.Set("type", customDomain.GetType()),
		d.Set("primary", customDomain.GetPrimary()),
		d.Set("status", customDomain.GetStatus()),
		d.Set("origin_domain_name", customDomain.GetOriginDomainName()),
	)

	if customDomain.Verification != nil {
		result = multierror.Append(result, d.Set("verification", []map[string]interface{}{
			{"methods": customDomain.Verification.Methods},
		}))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func deleteCustomDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if err := api.CustomDomain.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandCustomDomain(config cty.Value) *management.CustomDomain {
	return &management.CustomDomain{
		Domain: value.String(config.GetAttr("domain")),
		Type:   value.String(config.GetAttr("type")),
	}
}
