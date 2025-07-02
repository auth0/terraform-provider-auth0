package prompt

import (
	"encoding/json"
	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/value"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// Deprecated: expandPromptPartials is deprecated and will be removed in the next major version.
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
func expandPromptScreenPartial(data *schema.ResourceData) *management.PromptScreenPartials {
	partialRaw := data.GetRawConfig()
	if partialRaw.IsNull() {
		return nil
	}

	screenPartial := make(management.PromptScreenPartials)

	insertionPoints := expandInsertionPoints(partialRaw.GetAttr("insertion_points").AsValueSlice())
	screenPartial[management.ScreenName(partialRaw.GetAttr("screen_name").AsString())] = insertionPoints

	return &screenPartial
}

func expandPromptScreenPartials(data *schema.ResourceData) *management.PromptScreenPartials {
	partialsRaw := data.GetRawConfig().GetAttr("screen_partials")
	if partialsRaw.IsNull() {
		return nil
	}

	screenPartials := make(management.PromptScreenPartials)

	partialsRaw.ForEachElement(func(_ cty.Value, partialConfig cty.Value) (stop bool) {
		screenName := partialConfig.GetAttr("screen_name").AsString()
		insertionPoints := expandInsertionPoints(partialConfig.GetAttr("insertion_points").AsValueSlice())
		screenPartials[management.ScreenName(screenName)] = insertionPoints
		return stop
	})

	return &screenPartials
}

func expandInsertionPoints(insertionPointsList []cty.Value) map[management.InsertionPoint]string {
	insertionPoints := make(map[management.InsertionPoint]string)

	for _, insertionPoint := range insertionPointsList {
		insertionPointMap := insertionPoint.AsValueMap()

		if v := insertionPointMap["form_content"]; !v.IsNull() {
			insertionPoints[management.InsertionPointFormContent] = v.AsString()
		}
		if v := insertionPointMap["form_content_start"]; !v.IsNull() {
			insertionPoints[management.InsertionPointFormContentStart] = v.AsString()
		}
		if v := insertionPointMap["form_content_end"]; !v.IsNull() {
			insertionPoints[management.InsertionPointFormContentEnd] = v.AsString()
		}
		if v := insertionPointMap["form_footer_start"]; !v.IsNull() {
			insertionPoints[management.InsertionPointFormFooterStart] = v.AsString()
		}
		if v := insertionPointMap["form_footer_end"]; !v.IsNull() {
			insertionPoints[management.InsertionPointFormFooterEnd] = v.AsString()
		}
		if v := insertionPointMap["secondary_actions_start"]; !v.IsNull() {
			insertionPoints[management.InsertionPointSecondaryActionsStart] = v.AsString()
		}
		if v := insertionPointMap["secondary_actions_end"]; !v.IsNull() {
			insertionPoints[management.InsertionPointSecondaryActionsEnd] = v.AsString()
		}
	}

	return insertionPoints
}

func expandPromptSettings(data *schema.ResourceData) (*management.PromptRendering, error) {
	promptRawSettings := data.GetRawConfig()
	if promptRawSettings.IsNull() {
		return nil, nil
	}

	promptSettings := &management.PromptRendering{}

	promptSettings.RenderingMode = (*management.RenderingMode)(value.String(promptRawSettings.GetAttr("rendering_mode")))
	promptSettings.ContextConfiguration = value.Strings(promptRawSettings.GetAttr("context_configuration"))
	promptSettings.DefaultHeadTagsDisabled = value.Bool(promptRawSettings.GetAttr("default_head_tags_disabled"))
	if data.HasChange("head_tags") {
		promptSettings.HeadTags = expandInterfaceArray(data, "head_tags")
	}
	promptSettings.UsePageTemplate = value.Bool(promptRawSettings.GetAttr("use_page_template"))

	if data.HasChange("filters") {
		promptSettings.Filters = expandFilters(data)
	}

	return promptSettings, nil
}

func expandFilters(d *schema.ResourceData) *management.PromptRenderingFilters {
	filtersList := d.Get("filters").([]interface{})
	if len(filtersList) == 0 || filtersList[0] == nil {
		return nil
	}

	filterMap := filtersList[0].(map[string]interface{})
	return expandFilterStruct(filterMap)
}

func expandFilterStruct(m map[string]interface{}) *management.PromptRenderingFilters {
	f := &management.PromptRenderingFilters{}

	// match_type
	if v, ok := m["match_type"].(string); ok && v != "" {
		f.MatchType = auth0.String(v)
	}

	// clients
	if v, ok := m["clients"].(string); ok && v != "" {
		var clients []management.PromptRenderingFilter
		if err := json.Unmarshal([]byte(v), &clients); err == nil {
			f.Clients = &clients
		}
	}

	// organizations
	if v, ok := m["organizations"].(string); ok && v != "" {
		var orgs []management.PromptRenderingFilter
		if err := json.Unmarshal([]byte(v), &orgs); err == nil {
			f.Organizations = &orgs
		}
	}

	// domains
	if v, ok := m["domains"].(string); ok && v != "" {
		var domains []management.PromptRenderingFilter
		if err := json.Unmarshal([]byte(v), &domains); err == nil {
			f.Domains = &domains
		}
	}

	return f
}

//func expandPromptRenderingFilters(data *schema.ResourceData) *management.PromptRenderingFilters {
//	filtersInfo := expandInterfaceArray(data, "filters")
//	if filtersInfo == nil {
//		return nil
//	}
//
//	// Since filtersInfo is expected to be a list with a single object (map), extract the first
//	if len(filtersInfo) == 0 {
//		return nil
//	}
//
//	var filters management.PromptRenderingFilters
//
//	// Try again assuming it's a list with a single object
//	if m, ok := filtersInfo[0].(map[string]interface{}); ok {
//		b, err := json.Marshal(m)
//		if err != nil {
//			return nil
//		}
//		if err := json.Unmarshal(b, &filters); err != nil {
//			return nil
//		}
//		return &filters
//	}
//
//	return nil
//}

func expandInterfaceArray(d *schema.ResourceData, key string) []interface{} {
	_, newMetadata := d.GetChange(key)
	result := make([]interface{}, 0)
	if newMetadata == "" {
		return result
	}

	if newMetadataStr, ok := newMetadata.(string); ok {
		var newMetadataArr []interface{}
		if err := json.Unmarshal([]byte(newMetadataStr), &newMetadataArr); err != nil {
			return nil
		}
		return newMetadataArr
	}

	if newMetadataArr, ok := newMetadata.([]interface{}); ok {
		return newMetadataArr
	}

	return result
}

//func fetchNullableFields(data *schema.ResourceData) map[string]interface{} {
//	nullableMap := make(map[string]interface{})
//
//	for _, field := range []string{"filters", "head_tags"} {
//		if isFieldsNull(data, field) {
//			nullableMap[field] = nil
//		}
//	}
//
//	return nullableMap
//}
//
//func isFieldsNull(d *schema.ResourceData, fields string) bool {
//	raw, ok := d.GetOk(fields)
//	if !ok {
//		return true
//	}
//
//	str := strings.TrimSpace(raw.(string))
//	return str == "" || str == "null"
//}
