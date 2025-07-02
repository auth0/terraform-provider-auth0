package prompt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
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
	idComponents := strings.Split(data.Id(), ":")
	promptName, screenName := idComponents[0], idComponents[1]

	var insertionPoints interface{}
	if partial == nil || (*partial)[management.ScreenName(screenName)] == nil {
		insertionPoints = nil
	} else {
		insertionPoints = flattenInsertionPoints((*partial)[management.ScreenName(screenName)])
	}
	result := multierror.Append(
		data.Set("prompt_type", promptName),
		data.Set("screen_name", screenName),
		data.Set("insertion_points", insertionPoints),
	)
	return result.ErrorOrNil()
}

func flattenPromptScreenPartialsList(screenPartials *management.PromptScreenPartials) []map[string]interface{} {
	if screenPartials == nil {
		return nil
	}

	screenNames := make([]string, 0, len(*screenPartials))
	for screenName := range *screenPartials {
		screenNames = append(screenNames, string(screenName))
	}

	sort.Strings(screenNames)

	var screenPartialsList []map[string]interface{}

	for _, screenName := range screenNames {
		insertionPoints := (*screenPartials)[management.ScreenName(screenName)]
		flattenedInsertionPoints := flattenInsertionPoints(insertionPoints)

		screenPartialsList = append(screenPartialsList, map[string]interface{}{
			"screen_name":      screenName,
			"insertion_points": flattenedInsertionPoints,
		})
	}
	return screenPartialsList
}

func flattenInsertionPoints(insertionPoints map[management.InsertionPoint]string) []map[string]interface{} {
	if insertionPoints == nil {
		return nil
	}

	flattened := make(map[string]interface{})

	if v, exists := insertionPoints[management.InsertionPointFormContent]; exists {
		flattened["form_content"] = v
	}
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

func flattenPromptScreenSettings(data *schema.ResourceData, promptSetting *management.PromptRendering) error {
	var (
		idComponents           = strings.Split(data.Id(), ":")
		promptName, screenName = idComponents[0], idComponents[1]
	)

	result := multierror.Append(
		data.Set("prompt_type", promptName),
		data.Set("screen_name", screenName),
		data.Set("tenant", promptSetting.GetTenant()),
		data.Set("rendering_mode", promptSetting.GetRenderingMode()),
		data.Set("default_head_tags_disabled", promptSetting.GetDefaultHeadTagsDisabled()),
		data.Set("context_configuration", promptSetting.GetContextConfiguration()),
		data.Set("head_tags", flattenHeadTags(promptSetting)),
		data.Set("use_page_template", promptSetting.GetUsePageTemplate()),
		data.Set("filters", flattenPromptRenderingFilters(promptSetting.GetFilters())),
	)

	return result.ErrorOrNil()
}

func flattenHeadTags(promptSetting *management.PromptRendering) string {
	if promptSetting == nil || promptSetting.HeadTags == nil || len(promptSetting.HeadTags) == 0 {
		return ""
	}

	headTagBytes, err := json.Marshal(promptSetting.HeadTags)
	if err != nil {
		return ""
	}

	return string(headTagBytes)
}

func flattenPromptRenderingFilters(f *management.PromptRenderingFilters) []interface{} {
	if f == nil {
		return nil
	}

	return []interface{}{flattenFilter(f)}
}

func flattenFilter(f *management.PromptRenderingFilters) map[string]interface{} {

	result := make(map[string]interface{})

	// match_type
	if f.MatchType != nil {
		result["match_type"] = *f.MatchType
	}

	// clients
	if f.Clients != nil {
		if jsonStr, err := json.Marshal(f.Clients); err == nil {
			result["clients"] = string(jsonStr)
		}
	}

	// organizations
	if f.Organizations != nil {
		if jsonStr, err := json.Marshal(f.Organizations); err == nil {
			result["organizations"] = string(jsonStr)
		}
	}

	// domains
	if f.Domains != nil {
		if jsonStr, err := json.Marshal(f.Domains); err == nil {
			result["domains"] = string(jsonStr)
		}
	}

	return result
}

//
//func flattenPromptRenderingFilters(filters *management.PromptRenderingFilters) string {
//	if filters == nil {
//		return ""
//	}
//
//	out := make(map[string]interface{})
//
//	out["match_type"] = filters.GetMatchType()
//
//	if filters.Clients != nil {
//		out["clients"] = flattenPromptRenderingFilterList(filters.GetClients())
//	}
//
//	if filters.Organizations != nil {
//		out["organizations"] = flattenPromptRenderingFilterList(filters.GetOrganizations())
//	}
//
//	if filters.Domains != nil {
//		out["domains"] = flattenPromptRenderingFilterList(filters.GetDomains())
//	}
//
//	if len(out) == 0 {
//		return ""
//	}
//
//	// Wrap in list to match `jsonencode([filters])`
//	payload := []interface{}{out}
//
//	b, err := json.Marshal(payload)
//	if err != nil {
//		return ""
//	}
//
//	return string(b)
//}
//
//func flattenPromptRenderingFilterList(list []management.PromptRenderingFilter) []interface{} {
//	if list == nil || len(list) == 0 {
//		return nil
//	}
//
//	var result []interface{}
//	for _, f := range list {
//		item := make(map[string]interface{})
//
//		// Add "id" only if it's non-empty
//		if id := f.GetID(); id != "" {
//			item["id"] = id
//		}
//
//		// Add "metadata" only if it's non-nil and non-empty
//		if metadata := f.GetMetadata(); len(metadata) > 0 {
//			item["metadata"] = metadata
//		}
//
//		// Only add the item if it has at least one field (id or metadata)
//		if len(item) > 0 {
//			result = append(result, item)
//		}
//	}
//	return result
//}
