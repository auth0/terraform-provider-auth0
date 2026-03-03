package supplementalsignals

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_supplemental_signals resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createSupplementalSignals,
		UpdateContext: updateSupplementalSignals,
		ReadContext:   readSupplementalSignals,
		DeleteContext: deleteSupplementalSignals,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can configure Auth0 Supplemental Signals settings for your tenant. " +
			"This resource is a singleton, meaning only one instance exists per tenant.",
		Schema: map[string]*schema.Schema{
			"akamai_enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Indicates if incoming Akamai Headers should be processed.",
			},
		},
	}
}

func createSupplementalSignals(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId("supplemental_signals")

	return updateSupplementalSignals(ctx, data, meta)
}

func updateSupplementalSignals(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock("supplemental_signals")
	defer mutex.Unlock("supplemental_signals")

	apiv2 := meta.(*config.Config).GetAPIV2()

	supplementalSignalsConfig := expandSupplementalSignals(data)

	if _, err := apiv2.SupplementalSignals.Patch(ctx, supplementalSignalsConfig); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readSupplementalSignals(ctx, data, meta)
}

func readSupplementalSignals(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	supplementalSignalsConfig, err := apiv2.SupplementalSignals.Get(ctx)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId("supplemental_signals")

	return flattenSupplementalSignals(data, supplementalSignalsConfig)
}

func deleteSupplementalSignals(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock("supplemental_signals")
	defer mutex.Unlock("supplemental_signals")

	apiv2 := meta.(*config.Config).GetAPIV2()

	supplementalSignalsConfig := expandSupplementalSignalsForDelete()

	if _, err := apiv2.SupplementalSignals.Patch(ctx, supplementalSignalsConfig); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
