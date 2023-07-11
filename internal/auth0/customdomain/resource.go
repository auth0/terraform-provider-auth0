package customdomain

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewResource will return a new auth0_custom_domain resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCustomDomain,
		ReadContext:   readCustomDomain,
		UpdateContext: updateCustomDomain,
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
			"custom_client_ip_header": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"cf-connecting-ip", "x-forwarded-for", "true-client-ip", "",
				}, false),
				Description: "The HTTP header to fetch the client's IP address. " +
					"Cannot be set on auth0_managed domains.",
			},
			"tls_policy": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"compatible", "recommended",
				}, false),
				Description: "TLS policy for the custom domain. Available options are: `compatible` or `recommended`. " +
					"Compatible includes TLS 1.0, 1.1, 1.2, and recommended only includes TLS 1.2. " +
					"Cannot be set on self_managed domains.",
			},
		},
	}
}

func createCustomDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	customDomain := expandCustomDomain(d)
	if err := api.CustomDomain.Create(ctx, customDomain); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(customDomain.GetID())

	return readCustomDomain(ctx, d, m)
}

func readCustomDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	customDomain, err := api.CustomDomain.Read(ctx, d.Id())
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
		d.Set("custom_client_ip_header", customDomain.GetCustomClientIPHeader()),
		d.Set("tls_policy", customDomain.GetTLSPolicy()),
	)

	if customDomain.Verification != nil {
		result = multierror.Append(result, d.Set("verification", []map[string]interface{}{
			{"methods": customDomain.Verification.Methods},
		}))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateCustomDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	customDomain := expandCustomDomain(d)
	if err := api.CustomDomain.Update(ctx, d.Id(), customDomain); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return readCustomDomain(ctx, d, m)
}

func deleteCustomDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.CustomDomain.Delete(ctx, d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandCustomDomain(d *schema.ResourceData) *management.CustomDomain {
	config := d.GetRawConfig()

	customDomain := &management.CustomDomain{
		TLSPolicy:            value.String(config.GetAttr("tls_policy")),
		CustomClientIPHeader: value.String(config.GetAttr("custom_client_ip_header")),
	}

	if d.IsNewResource() {
		customDomain.Domain = value.String(config.GetAttr("domain"))
		customDomain.Type = value.String(config.GetAttr("type"))
	}

	return customDomain
}
