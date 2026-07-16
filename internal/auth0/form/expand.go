package form

import (
	"encoding/json"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandForm(data *schema.ResourceData) (*management.Form, error) {
	cfg := data.GetRawConfig()

	form := &management.Form{}

	form.Name = value.String(cfg.GetAttr("name"))
	form.Languages = expandFomLanguages(cfg.GetAttr("languages"))

	if data.HasChange("messages") {
		form.Messages = expandFomMessages(cfg.GetAttr("messages"))
	}

	if data.HasChange("translations") {
		translations, err := expandStringInterfaceMap(data, "translations")
		if err != nil {
			return nil, err
		}
		form.Translations = &translations
	}

	if data.HasChange("style") {
		style, err := expandStringInterfaceMap(data, "style")
		if err != nil {
			return nil, err
		}
		form.Style = &style
	}

	if data.HasChange("start") {
		start, err := expandStringInterfaceMap(data, "start")
		if err != nil {
			return nil, err
		}
		form.Start = &start
	}

	if data.HasChange("ending") {
		ending, err := expandStringInterfaceMap(data, "ending")
		if err != nil {
			return nil, err
		}
		form.Ending = &ending
	}

	if data.HasChange("nodes") {
		form.Nodes = expandInterfaceArray(data, "nodes")
	}

	return form, nil
}

func expandFomLanguages(languages cty.Value) *management.FormLanguages {
	if languages.IsNull() {
		return nil
	}

	formLanguages := &management.FormLanguages{}

	languages.ForEachElement(func(_ cty.Value, language cty.Value) (stop bool) {
		formLanguages.Primary = value.String(language.GetAttr("primary"))
		formLanguages.Default = value.String(language.GetAttr("default"))
		return stop
	})

	return formLanguages
}

func expandFomMessages(messages cty.Value) *management.FormMessages {
	if messages.IsNull() {
		return nil
	}

	formMessages := &management.FormMessages{}

	messages.ForEachElement(func(_ cty.Value, message cty.Value) (stop bool) {
		if message.GetAttr("custom").IsNull() {
			formMessages.Custom = nil
		} else {
			formMessages.Custom = convertToInterfaceMap(message.GetAttr("custom"))
		}

		if message.GetAttr("errors").IsNull() {
			formMessages.Errors = nil
		} else {
			formMessages.Errors = convertToInterfaceMap(message.GetAttr("errors"))
		}

		return stop
	})

	return formMessages
}

func expandInterfaceArray(d *schema.ResourceData, key string) []interface{} {
	oldMetadata, newMetadata := d.GetChange(key)
	result := make([]interface{}, 0)
	if oldMetadata == "" && newMetadata == "" {
		return result
	}

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

func expandStringInterfaceMap(data *schema.ResourceData, key string) (map[string]interface{}, error) {
	oldMetadata, newMetadata := data.GetChange(key)
	if oldMetadata == "" {
		return value.MapFromJSON(data.GetRawConfig().GetAttr(key))
	}

	if newMetadata == "" {
		return map[string]interface{}{}, nil
	}

	oldMap, err := structure.ExpandJsonFromString(oldMetadata.(string))
	if err != nil {
		return map[string]interface{}{}, err
	}

	newMap, err := structure.ExpandJsonFromString(newMetadata.(string))
	if err != nil {
		return map[string]interface{}{}, err
	}

	for key := range oldMap {
		if _, ok := newMap[key]; !ok {
			newMap[key] = nil
		}
	}

	return newMap, nil
}

func convertToInterfaceMap(rawValue cty.Value) *map[string]interface{} {
	if rawValue.IsNull() {
		return nil
	}

	m := make(map[string]interface{})
	m, err := structure.ExpandJsonFromString(rawValue.AsString())
	if err != nil {
		return &m
	}

	return &m
}

func isNodeEmpty(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("nodes") {
		return false
	}
	nodes := expandInterfaceArray(data, "nodes")
	return len(nodes) == 0
}
