package prompt

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenPrompt(data *schema.ResourceData, prompt *management.Prompt) error {
	result := multierror.Append(
		data.Set("universal_login_experience", prompt.UniversalLoginExperience),
		data.Set("identifier_first", prompt.GetIdentifierFirst()),
		data.Set("webauthn_platform_first_factor", prompt.GetWebAuthnPlatformFirstFactor()),
	)

	return result.ErrorOrNil()
}

func flattenPromptPartials(data *schema.ResourceData, promptPartials *management.PromptScreenPartials) error {
	var result = &multierror.Error{}

	screenName := management.ScreenName(data.Get("screen_name").(string))
	if screenName == "" {
		screenName = management.ScreenName(data.Get("prompt").(string))
	}

	if insertionPoints, ok := (*promptPartials)[screenName]; ok {
		for insertionPoint, content := range insertionPoints {
			switch insertionPoint {
			case management.InsertionPointFormContentStart:
				result = multierror.Append(result, data.Set("form_content_start", content))
			case management.InsertionPointFormContentEnd:
				result = multierror.Append(result, data.Set("form_content_end", content))
			case management.InsertionPointFormFooterStart:
				result = multierror.Append(result, data.Set("form_footer_start", content))
			case management.InsertionPointFormFooterEnd:
				result = multierror.Append(result, data.Set("form_footer_end", content))
			case management.InsertionPointSecondaryActionsStart:
				result = multierror.Append(result, data.Set("secondary_actions_start", content))
			case management.InsertionPointSecondaryActionsEnd:
				result = multierror.Append(result, data.Set("secondary_actions_end", content))
			default:
				result = multierror.Append(result, fmt.Errorf("unknown insertion point %q for screen %q", insertionPoint, screenName))
			}
		}
	} else {
		result = multierror.Append(result, fmt.Errorf("screen %q not found in prompt partials", screenName))
	}

	return result.ErrorOrNil()
}

func flattenPromptCustomText(data *schema.ResourceData, customText map[string]interface{}) error {
	body, err := marshalCustomTextBody(customText)
	if err != nil {
		return err
	}

	return data.Set("body", body)
}

func marshalCustomTextBody(b map[string]interface{}) (string, error) {
	if b == nil {
		return "{}", nil
	}

	bodyBytes, err := json.Marshal(b)
	if err != nil {
		return "", fmt.Errorf("failed to serialize the custom texts to JSON: %w", err)
	}

	var buffer bytes.Buffer
	const jsonIndentation = "    "
	if err := json.Indent(&buffer, bodyBytes, "", jsonIndentation); err != nil {
		return "", fmt.Errorf("failed to format the custom texts JSON: %w", err)
	}

	return buffer.String(), nil
}
