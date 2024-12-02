package selfserviceprofile

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewCustomTextResource will return a new auth0_self_service_profile_custom_text resource.
func NewCustomTextResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCustomTextForSSOProfile,
		ReadContext:   readCustomTextForSSOProfile,
		UpdateContext: updateCustomTextForSSOProfile,
		DeleteContext: deleteCustomTextForSSOProfile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can set custom text for Self-Service Profile",
		Schema: map[string]*schema.Schema{
			"sso_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The id of the self-service profile",
			},
			"language": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The language of the custom text",
			},
			"page": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The page where the custom text is shown",
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description: "The list of text keys and values to customize the self-service SSO page. " +
					"Values can be plain text or rich HTML content limited to basic styling tags and hyperlinks",
			},
		},
	}
}

func createCustomTextForSSOProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := data.Get("sso_id").(string)
	language := data.Get("language").(string)
	page := data.Get("page").(string)

	internalSchema.SetResourceGroupID(data, id, language, page)

	return updateCustomTextForSSOProfile(ctx, data, meta)
}

func readCustomTextForSSOProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customText, err := api.SelfServiceProfile.GetCustomText(ctx,
		data.Get("sso_id").(string),
		data.Get("language").(string),
		data.Get("page").(string))
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenSSOCustomText(data, customText))
}

func updateCustomTextForSSOProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	id := data.Get("sso_id").(string)
	language := data.Get("language").(string)
	page := data.Get("page").(string)
	body := data.Get("body").(string)

	if body == "" {
		return nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return diag.FromErr(err)
	}

	if err := api.SelfServiceProfile.SetCustomText(ctx, id, language, page, payload); err != nil {
		return diag.FromErr(err)
	}

	return readCustomTextForSSOProfile(ctx, data, meta)
}

func deleteCustomTextForSSOProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := data.Set("body", "{}"); err != nil {
		return diag.FromErr(err)
	}

	return updateCustomTextForSSOProfile(ctx, data, meta)
}
