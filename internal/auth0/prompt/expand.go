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

func expandPromptPartials(data *schema.ResourceData) *management.PromptScreenPartials {
	screenName := management.ScreenName(data.Get("screen_name").(string))
	if screenName == "" {
		screenName = management.ScreenName(data.Get("prompt").(string))
	}

	insertionPoints := make(map[management.InsertionPoint]string)

	if content := data.Get("form_content_start").(string); content != "" {
		insertionPoints[management.InsertionPointFormContentStart] = content
	}
	if content := data.Get("form_content_end").(string); content != "" {
		insertionPoints[management.InsertionPointFormContentEnd] = content
	}
	if content := data.Get("form_footer_start").(string); content != "" {
		insertionPoints[management.InsertionPointFormFooterStart] = content
	}
	if content := data.Get("form_footer_end").(string); content != "" {
		insertionPoints[management.InsertionPointFormFooterEnd] = content
	}
	if content := data.Get("secondary_actions_start").(string); content != "" {
		insertionPoints[management.InsertionPointSecondaryActionsStart] = content
	}
	if content := data.Get("secondary_actions_end").(string); content != "" {
		insertionPoints[management.InsertionPointSecondaryActionsEnd] = content
	}

	return &management.PromptScreenPartials{
		screenName: insertionPoints,
	}
}
