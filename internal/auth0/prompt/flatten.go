package prompt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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

// Deprecated: flattenPromptPartials is deprecated and will be removed in the next major version.
func flattenPromptPartials(data *schema.ResourceData, promptPartials *management.PromptPartials) error {
	result := multierror.Append(
		data.Set("form_content_start", promptPartials.FormContentStart),
		data.Set("form_content_end", promptPartials.FormContentEnd),
		data.Set("form_footer_start", promptPartials.FormFooterStart),
		data.Set("form_footer_end", promptPartials.FormFooterEnd),
		data.Set("secondary_actions_start", promptPartials.SecondaryActionsStart),
		data.Set("secondary_actions_end", promptPartials.SecondaryActionsEnd),
	)

	return result.ErrorOrNil()
}

func flattenPromptScreenPartials(data *schema.ResourceData, screenPartials *management.PromptScreenPartials) error {
	result := multierror.Append(
		data.Set("prompt_type", data.Id()),
		data.Set("screen_partials", flattenPromptScreenPartialsList(screenPartials)),
	)
	return result.ErrorOrNil()
}

func flattenPromptScreenPartial(data *schema.ResourceData, partial *management.PromptScreenPartials) error {
	promptName, screenName := strings.Split(data.Id(), ":")[0], strings.Split(data.Id(), ":")[1]
	result := multierror.Append(
		data.Set("prompt_type", promptName),
		data.Set("screen_name", screenName),
		data.Set("insertion_points", flattenInsertionPoints((*partial)[management.ScreenName(screenName)])),
	)
	return result.ErrorOrNil()
}

func flattenPromptScreenPartialsList(screenPartials *management.PromptScreenPartials) []map[string]interface{} {
	if screenPartials == nil {
		return nil
	}

	var screenPartialsList []map[string]interface{}

	for screenName, insertionPoints := range *screenPartials {
		flattenedInsertionPoints := flattenInsertionPoints(insertionPoints)

		screenPartialsList = append(screenPartialsList, map[string]interface{}{
			"screen_name":      string(screenName),
			"insertion_points": flattenedInsertionPoints, // This should now be a []map[string]interface{}.
		})
	}
	return screenPartialsList
}

func flattenInsertionPoints(insertionPoints map[management.InsertionPoint]string) []map[string]interface{} {
	if insertionPoints == nil {
		return nil
	}

	flattened := make(map[string]interface{})

	if v, exists := insertionPoints[management.InsertionPointFormContentStart]; exists {
		flattened["form_content_start"] = v
	}
	if v, exists := insertionPoints[management.InsertionPointFormContentEnd]; exists {
		flattened["form_content_end"] = v
	}
	if v, exists := insertionPoints[management.InsertionPointFormFooterStart]; exists {
		flattened["form_footer_start"] = v
	}
	if v, exists := insertionPoints[management.InsertionPointFormFooterEnd]; exists {
		flattened["form_footer_end"] = v
	}
	if v, exists := insertionPoints[management.InsertionPointSecondaryActionsStart]; exists {
		flattened["secondary_actions_start"] = v
	}
	if v, exists := insertionPoints[management.InsertionPointSecondaryActionsEnd]; exists {
		flattened["secondary_actions_end"] = v
	}

	return []map[string]interface{}{flattened}
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
