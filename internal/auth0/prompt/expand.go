package prompt

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandPrompt(data cty.Value) *management.Prompt {
	prompt := management.Prompt{
		IdentifierFirst:             value.Bool(data.GetAttr("identifier_first")),
		WebAuthnPlatformFirstFactor: value.Bool(data.GetAttr("webauthn_platform_first_factor")),
	}

	ulExp := data.GetAttr("universal_login_experience")
	if !ulExp.IsNull() {
		prompt.UniversalLoginExperience = ulExp.AsString()
	}

	return &prompt
}

func expandPromptPartials(data *schema.ResourceData) *management.PromptPartials {
	return &management.PromptPartials{
		Prompt:                management.PromptType(data.Get("prompt").(string)),
		FormContentStart:      data.Get("form_content_start").(string),
		FormContentEnd:        data.Get("form_content_end").(string),
		FormFooterStart:       data.Get("form_footer_start").(string),
		FormFooterEnd:         data.Get("form_footer_end").(string),
		SecondaryActionsStart: data.Get("secondary_actions_start").(string),
		SecondaryActionsEnd:   data.Get("secondary_actions_end").(string),
	}
}
