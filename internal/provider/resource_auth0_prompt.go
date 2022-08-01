package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newPrompt() *schema.Resource {
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
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"new", "classic",
				}, false),
				Description: "Which login experience to use. Options include `classic` and `new`.",
			},
			"identifier_first": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Indicates whether the identifier first is used when " +
					"using the new universal login experience.",
			},
			"webauthn_platform_first_factor": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Determines if the login screen uses identifier and biometrics first.",
			},
		},
	}
}

func createPrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(resource.UniqueId())
	return updatePrompt(ctx, d, m)
}

func readPrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	prompt, err := api.Prompt.Read()
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("universal_login_experience", prompt.UniversalLoginExperience),
		d.Set("identifier_first", prompt.IdentifierFirst),
		d.Set("webauthn_platform_first_factor", prompt.WebAuthnPlatformFirstFactor),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updatePrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	prompt := buildPrompt(d)
	api := m.(*management.Management)
	if err := api.Prompt.Update(prompt); err != nil {
		return diag.FromErr(err)
	}

	return readPrompt(ctx, d, m)
}

func deletePrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func buildPrompt(d *schema.ResourceData) *management.Prompt {
	return &management.Prompt{
		UniversalLoginExperience:    auth0.StringValue(String(d, "universal_login_experience")),
		IdentifierFirst:             Bool(d, "identifier_first"),
		WebAuthnPlatformFirstFactor: Bool(d, "webauthn_platform_first_factor"),
	}
}
