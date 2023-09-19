package prompt

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewResource will return a new auth0_prompt resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPrompt,
		ReadContext:   readPrompt,
		UpdateContext: updatePrompt,
		DeleteContext: deletePrompt,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage your Auth0 prompts, " +
			"including choosing the login experience version.",
		Schema: map[string]*schema.Schema{
			"universal_login_experience": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"new", "classic"}, false),
				AtLeastOneOf: []string{"identifier_first", "webauthn_platform_first_factor"},
				Description:  "Which login experience to use. Options include `classic` and `new`.",
			},
			"identifier_first": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"universal_login_experience", "webauthn_platform_first_factor"},
				Description: "Indicates whether the identifier first is used when " +
					"using the new Universal Login experience.",
			},
			"webauthn_platform_first_factor": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"universal_login_experience", "identifier_first"},
				Description: "Determines if the login screen uses identifier and biometrics first. " +
					"Setting this property to `true`, requires MFA factors enabled for enrollment; use the `auth0_guardian` resource to set one up.",
			},
		},
	}
}

func createPrompt(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updatePrompt(ctx, data, meta)
}

func readPrompt(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	prompt, err := api.Prompt.Read(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenPrompt(data, prompt))
}

func updatePrompt(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	prompt := expandPrompt(data.GetRawConfig())

	if err := api.Prompt.Update(ctx, prompt); err != nil {
		return diag.FromErr(err)
	}

	return readPrompt(ctx, data, meta)
}

func deletePrompt(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
