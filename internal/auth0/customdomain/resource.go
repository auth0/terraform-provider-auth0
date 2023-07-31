package customdomain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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

func createCustomDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customDomain := expandCustomDomain(data)

	if err := api.CustomDomain.Create(ctx, customDomain); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(customDomain.GetID())

	return readCustomDomain(ctx, data, meta)
}

func readCustomDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customDomain, err := api.CustomDomain.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenCustomDomain(data, customDomain))
}

func updateCustomDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customDomain := expandCustomDomain(data)

	if err := api.CustomDomain.Update(ctx, data.Id(), customDomain); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readCustomDomain(ctx, data, meta)
}

func deleteCustomDomain(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.CustomDomain.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
