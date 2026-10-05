package phoneproviderprotection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_phone_provider_protection resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPhoneProviderProtection,
		UpdateContext: updatePhoneProviderProtection,
		ReadContext:   readPhoneProviderProtection,
		DeleteContext: deletePhoneProviderProtection,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can configure the MFA enrollment backoff strategy for phone-based " +
			"multi-factor authentication on your Auth0 tenant. " +
			"This resource is a singleton, meaning only one instance exists per tenant.",
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"exponential",
					"default",
				}, false),
				Description: "The backoff strategy for MFA enrollment. " +
					"Allowed values are `exponential` and `default`.",
			},
		},
	}
}

func createPhoneProviderProtection(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("phone_provider_protection")

	return updatePhoneProviderProtection(ctx, d, meta)
}

func readPhoneProviderProtection(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	cfg, err := apiv3.AttackProtection.PhoneProviderProtection.Get(ctx)
	if err != nil {
		if internalError.IsInsufficientEntitlement(err) {
			return diag.Diagnostics{internalError.EntitlementWarning("Phone Provider Protection", internalError.EntitlementReadConsequence, err)}
		}
		return internalError.HandleReadAPIError("auth0_phone_provider_protection", d, err)
	}

	d.SetId("phone_provider_protection")

	return flattenPhoneProviderProtection(d, cfg)
}

func updatePhoneProviderProtection(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock("phone_provider_protection")
	defer mutex.Unlock("phone_provider_protection")

	apiv3 := meta.(*config.Config).GetAPIV3()

	cfg := expandPhoneProviderProtection(d)

	var diags diag.Diagnostics

	if _, err := apiv3.AttackProtection.PhoneProviderProtection.Patch(ctx, cfg); err != nil {
		if internalError.IsInsufficientEntitlement(err) {
			diags = append(diags, internalError.EntitlementWarning("Phone Provider Protection", internalError.EntitlementUpdateConsequence, err))
		} else {
			return diag.FromErr(internalError.HandleAPIError(d, err))
		}
	}

	return append(diags, readPhoneProviderProtection(ctx, d, meta)...)
}

func deletePhoneProviderProtection(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock("phone_provider_protection")
	defer mutex.Unlock("phone_provider_protection")

	apiv3 := meta.(*config.Config).GetAPIV3()

	cfg := expandPhoneProviderProtectionForDelete()

	if _, err := apiv3.AttackProtection.PhoneProviderProtection.Patch(ctx, cfg); err != nil {
		if internalError.IsInsufficientEntitlement(err) {
			return diag.Diagnostics{internalError.EntitlementWarning("Phone Provider Protection", internalError.EntitlementUpdateConsequence, err)}
		}
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return nil
}
