package branding

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var supportedDeliveryMethods = []string{"voice", "text"}

// NewPhoneProviderResource will return a new auth0_phone_provider resource.
func NewPhoneProviderResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPhoneProvider,
		ReadContext:   readPhoneProvider,
		UpdateContext: updatePhoneProvider,
		DeleteContext: deletePhoneProvider,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Auth0 allows you to configure your own phone messaging provider to help you manage, monitor, " +
			"and troubleshoot your SMS and voice communications. You can only configure one phone provider for all SMS " +
			"and voice communications per tenant.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"twilio", "custom"},
					false,
				),
				Description: "Name of the phone provider. Options include `twilio`, `custom`.",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the phone provider is enabled (false) or disabled (true).",
			},
			"channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The channel of the phone provider.",
			},
			"tenant": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tenant of the phone provider.",
			},
			"credentials": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Provider credentials required to use authenticate to the provider.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_token": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Sensitive:    true,
							Description:  "The Auth Token for the phone provider.",
						},
					},
				},
				DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
					providerName := d.Get("name").(string) // Get another field's value.
					// For "twilio", allow changes only if the value has actually changed.
					if providerName == "twilio" && d.HasChange("auth_token") {
						return false // Allow the diff if the auth_token has actually changed.
					}

					return true // Allow changes otherwise.
				},
			},
			"configuration": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Specific phone provider settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delivery_methods": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Media set supported by a given provider to deliver a notification",
							MaxItems:    2,
							MinItems:    1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(supportedDeliveryMethods, false),
							},
						},
						"default_from": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Default sender subject as \"from\" when no other value is specified.",
							RequiredWith: []string{"configuration.0.sid"},
						},
						"mssid": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Twilio Messaging Service SID",
						},
						"sid": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Twilio Account SID.",
							RequiredWith: []string{"configuration.0.default_from"},
						},
					},
				},
			},
		},
	}
}

func createPhoneProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if id, isConfigured := phoneProviderIsConfigured(ctx, api); isConfigured {
		data.SetId(id)
		return updatePhoneProvider(ctx, data, meta)
	}

	phoneProviderConfig := expandPhoneProvider(data.GetRawConfig())

	if err := api.Branding.CreatePhoneProvider(ctx, phoneProviderConfig); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(phoneProviderConfig.GetID())

	return readPhoneProvider(ctx, data, meta)
}

func readPhoneProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	phoneProviderConfig, err := api.Branding.ReadPhoneProvider(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenPhoneProvider(data, phoneProviderConfig))
}

func updatePhoneProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	phoneProviderConfig := expandPhoneProvider(data.GetRawConfig())

	if err := api.Branding.UpdatePhoneProvider(ctx, data.Id(), phoneProviderConfig); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readPhoneProvider(ctx, data, meta)
}

func deletePhoneProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Branding.DeletePhoneProvider(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func phoneProviderIsConfigured(ctx context.Context, api *management.Management) (string, bool) {
	phoneProviders, _ := api.Branding.ListPhoneProviders(ctx)

	if len(phoneProviders.Providers) > 0 {
		return phoneProviders.Providers[0].GetID(), true
	}

	return "", false
}
