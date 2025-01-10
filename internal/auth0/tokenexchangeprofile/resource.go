package tokenexchangeprofile

import (
	"context"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewResource will return a new auth0_token_exchange_profile resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createTokenExchangeProfile,
		ReadContext:   readTokenExchangeProfile,
		UpdateContext: updateTokenExchangeProfile,
		DeleteContext: deleteTokenExchangeProfile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage Auth0 Custom Token Exchange Profiles",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the token exchange profile.",
			},
			"subject_token_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of the subject token",
			},
			"action_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Unique identifier of the Action",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Type of the token exchange profile",
			},
		},
	}
}

func createTokenExchangeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	tep := expandTokenExchangeProfiles(data)

	if err := api.TokenExchangeProfile.Create(ctx, tep); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(tep.GetID())

	return readTokenExchangeProfile(ctx, data, meta)
}

func readTokenExchangeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	ssp, err := api.TokenExchangeProfile.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenTokenExchangeProfile(data, ssp))
}

func updateTokenExchangeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	tep := expandTokenExchangeProfiles(data)

	if err := api.TokenExchangeProfile.Update(ctx, data.Id(), tep); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readTokenExchangeProfile(ctx, data, meta)
}

func deleteTokenExchangeProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.TokenExchangeProfile.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
