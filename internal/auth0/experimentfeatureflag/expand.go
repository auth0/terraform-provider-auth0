package experimentfeatureflag

import (
	"encoding/json"
	"fmt"
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
	var hasChanged bool
	cfg := data.GetRawConfig()

	request := &management.UpdateFeatureFlagRequestContent{}

	// Only send fields the user changed. The Set* helpers mark each field explicit so a
	// cleared optional (e.g. a removed description) serializes as null instead of being omitted.
	if data.HasChange("name") {
		request.SetName(value.String(cfg.GetAttr("name")))
		hasChanged = true
	}

	if data.HasChange("description") {
		request.SetDescription(value.String(cfg.GetAttr("description")))
		hasChanged = true
	}

	if data.HasChange("parameters") {
		parameters, err := expandParameters(data.Get("parameters").(*schema.Set))
		if err != nil {
			return nil, err
		}

		request.SetParameters(&parameters)
		hasChanged = true
	}

	if !hasChanged {
		return nil, nil
	}

	return request, nil
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

// expandVariations reads the nested variation blocks into an intermediate form used by both
// create and reconcile.
func expandVariations(data *schema.ResourceData) ([]variationConfig, error) {
	paramTypesByName := parameterTypesByName(data.Get("parameters").(*schema.Set))

	rawVariations := data.Get("variation").([]interface{})
	variations := make([]variationConfig, 0, len(rawVariations))

	for _, raw := range rawVariations {
		cfg, err := expandVariation(raw.(map[string]interface{}), paramTypesByName)
		if err != nil {
			return nil, err
		}

		variations = append(variations, cfg)
	}

	return variations, nil
}

// expandVariation converts a single nested variation block into a variationConfig.
func expandVariation(entry map[string]interface{}, paramTypesByName map[string]string) (variationConfig, error) {
	name := entry["name"].(string)

	overrides, err := expandOverrideValues(entry["overrides"].(map[string]interface{}), paramTypesByName)
	if err != nil {
		return variationConfig{}, fmt.Errorf("variation %q: %w", name, err)
	}

	cfg := variationConfig{name: name, overrides: overrides}
	// Keep empty descriptions nil: the API rejects `""` (min length 3), so create omits the field
	// and update sends null (`SetDescription(nil)`) to clear an existing value.
	if description, ok := entry["description"].(string); ok && description != "" {
		cfg.description = &description
	}

	return cfg, nil
}

func expandOverrideValues(raw map[string]interface{}, paramTypesByName map[string]string) (map[string]interface{}, error) {
	overrides := make(map[string]interface{}, len(raw))

	for name, value := range raw {
		raw := value.(string)

		typed, err := stringToTyped(paramTypesByName[name], raw)
		if err != nil {
			return nil, fmt.Errorf("override %q: %w", name, err)
		}
		overrides[name] = typed
	}

	return overrides, nil
}

func parameterTypesByName(set *schema.Set) map[string]string {
	types := make(map[string]string, set.Len())
	for _, raw := range set.List() {
		entry := raw.(map[string]interface{})
		types[entry["name"].(string)] = entry["type"].(string)
	}
	return types
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
		f, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return nil, fmt.Errorf("value %q is not a valid number", rawValue)
		}
		return f, nil
	case management.FeatureFlagConfigParamTypeEnumObject, management.FeatureFlagConfigParamTypeEnumArray:
		var obj interface{}
		if err := json.Unmarshal([]byte(rawValue), &obj); err != nil {
			return nil, fmt.Errorf("value %q is not valid JSON for %s", rawValue, paramType)
		}
		return obj, nil
	default:
		return rawValue, nil
	}
}

// variationConfig is the parsed configuration of a single nested variation block.
type variationConfig struct {
	name        string
	description *string
	overrides   map[string]interface{}
}

func (cfg variationConfig) createRequest() *management.CreateVariationRequestContent {
	overrides := make(management.VariationOverridesMap, len(cfg.overrides))
	for name, value := range cfg.overrides {
		overrides[name] = value
	}

	return &management.CreateVariationRequestContent{
		Name:        cfg.name,
		Overrides:   overrides,
		Description: cfg.description,
	}
}

func (cfg variationConfig) updateRequest() *management.UpdateVariationRequestContent {
	overrides := make(management.UpdateVariationOverridesMap, len(cfg.overrides))
	for name, value := range cfg.overrides {
		overrides[name] = value
	}

	request := &management.UpdateVariationRequestContent{
		Name:      auth0.String(cfg.name),
		Overrides: &overrides,
	}
	request.SetDescription(cfg.description)

	return request
}
