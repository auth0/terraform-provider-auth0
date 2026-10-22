package experimentfeatureflag

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/auth0/go-auth0"
	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandFeatureFlag(data *schema.ResourceData) (*management.CreateFeatureFlagRequestContent, error) {
	parameters, err := expandParameters(data.Get("parameters").(*schema.Set))
	if err != nil {
		return nil, err
	}

	cfg := data.GetRawConfig()

	request := &management.CreateFeatureFlagRequestContent{
		Name:        auth0.StringValue(value.String(cfg.GetAttr("name"))),
		Description: value.String(cfg.GetAttr("description")),
		Parameters:  parameters,
	}

	return request, nil
}

func expandFeatureFlagUpdate(data *schema.ResourceData) (*management.UpdateFeatureFlagRequestContent, error) {
	cfg := data.GetRawConfig()

	request := &management.UpdateFeatureFlagRequestContent{}

	// Only send fields the user changed. The Set* helpers mark each field explicit so a
	// cleared optional (e.g. a removed description) serializes as null instead of being omitted.
	if data.HasChange("name") {
		request.SetName(value.String(cfg.GetAttr("name")))
	}

	if data.HasChange("description") {
		request.SetDescription(value.String(cfg.GetAttr("description")))
	}

	if data.HasChange("parameters") {
		parameters, err := expandParameters(data.Get("parameters").(*schema.Set))
		if err != nil {
			return nil, err
		}
		request.SetParameters(&parameters)
	}

	return request, nil
}

// expandFeatureFlagStatusOnCreate returns nil when the status is omitted or draft,
// because a new feature flag is created as a draft.
func expandFeatureFlagStatusOnCreate(data *schema.ResourceData) *management.FeatureFlagStatusEnum {
	status := value.String(data.GetRawConfig().GetAttr("status"))
	if status == nil || *status == string(management.FeatureFlagStatusEnumDraft) {
		return nil
	}

	typed := management.FeatureFlagStatusEnum(*status)
	return &typed
}

func expandParameters(set *schema.Set) (map[string]*management.FeatureFlagConfigParam, error) {
	parameters := make(map[string]*management.FeatureFlagConfigParam, set.Len())

	for _, raw := range set.List() {
		entry := raw.(map[string]interface{})
		name := entry["name"].(string)
		paramType := entry["type"].(string)

		value, err := stringToTyped(paramType, entry["value"].(string))
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}

		param := &management.FeatureFlagConfigParam{
			Type:  management.FeatureFlagConfigParamTypeEnum(paramType),
			Value: value,
		}
		if description, ok := entry["description"].(string); ok && description != "" {
			param.Description = auth0.String(description)
		}

		parameters[name] = param
	}

	return parameters, nil
}

// stringToTyped converts a Terraform string value into the raw typed value the API expects
// for the given parameter type. The type must be one of the known parameter types.
func stringToTyped(paramType, rawValue string) (interface{}, error) {
	typed := management.FeatureFlagConfigParamTypeEnum(paramType)

	switch typed {
	case management.FeatureFlagConfigParamTypeEnumBoolean:
		b, err := strconv.ParseBool(rawValue)
		if err != nil {
			return nil, fmt.Errorf("value %q is not a valid boolean", rawValue)
		}
		return b, nil
	case management.FeatureFlagConfigParamTypeEnumNumber:
		// ParseFloat accepts NaN and Inf, which JSON cannot encode.
		f, err := strconv.ParseFloat(rawValue, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("value %q is not a valid finite number", rawValue)
		}
		return f, nil
	case management.FeatureFlagConfigParamTypeEnumObject:
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(rawValue), &obj); err != nil || obj == nil {
			return nil, fmt.Errorf("value %q is not a JSON object", rawValue)
		}
		return obj, nil
	case management.FeatureFlagConfigParamTypeEnumArray:
		var arr []interface{}
		if err := json.Unmarshal([]byte(rawValue), &arr); err != nil || arr == nil {
			return nil, fmt.Errorf("value %q is not a JSON array", rawValue)
		}
		return arr, nil
	default:
		return rawValue, nil
	}
}

// expandVariation reads an empty description as unset, because a set element
// cannot tell an empty string from an omitted attribute.
func expandVariation(variation map[string]interface{}) (*management.CreateVariationRequestContent, error) {
	overrides, err := expandOverrides(variation["overrides"].(*schema.Set))
	if err != nil {
		return nil, err
	}

	request := &management.CreateVariationRequestContent{
		Name:      variation["name"].(string),
		Overrides: overrides,
	}
	if description := variation["description"].(string); description != "" {
		request.Description = auth0.String(description)
	}

	return request, nil
}

// expandVariationUpdate compares two variations with the same name and returns
// false when nothing changed. A removed description is sent as null to clear it.
func expandVariationUpdate(oldVariation, newVariation map[string]interface{}) (*management.UpdateVariationRequestContent, bool, error) {
	request := &management.UpdateVariationRequestContent{}
	changed := false

	if oldDescription, newDescription := oldVariation["description"].(string), newVariation["description"].(string); oldDescription != newDescription {
		var description *string
		if newDescription != "" {
			description = auth0.String(newDescription)
		}
		request.SetDescription(description)
		changed = true
	}

	oldOverrides, newOverrides := oldVariation["overrides"].(*schema.Set), newVariation["overrides"].(*schema.Set)
	if !oldOverrides.HashEqual(newOverrides) {
		overrides, err := expandOverridesUpdate(oldOverrides, newOverrides)
		if err != nil {
			return nil, false, err
		}
		request.SetOverrides(&overrides)
		changed = true
	}

	return request, changed, nil
}

// expandOverridesUpdate builds the PATCH overrides. PATCH merges override keys,
// so the full new set is sent and each removed key is sent as null to delete it.
// An empty object is never sent, so removing every override also uses null keys.
func expandOverridesUpdate(oldSet, newSet *schema.Set) (map[string]interface{}, error) {
	overrides, err := expandOverrides(newSet)
	if err != nil {
		return nil, err
	}

	for _, name := range overrideNames(oldSet) {
		if _, ok := overrides[name]; !ok {
			overrides[name] = nil
		}
	}

	return overrides, nil
}

// expandOverrides always returns a non-nil map, because POST requires
// overrides and an empty map creates a control variation.
func expandOverrides(set *schema.Set) (map[string]interface{}, error) {
	overrides := make(map[string]interface{}, set.Len())

	for _, raw := range set.List() {
		entry := raw.(map[string]interface{})
		name := entry["name"].(string)

		value, err := stringToTyped(entry["type"].(string), entry["value"].(string))
		if err != nil {
			return nil, fmt.Errorf("override %q: %w", name, err)
		}

		overrides[name] = value
	}

	return overrides, nil
}
