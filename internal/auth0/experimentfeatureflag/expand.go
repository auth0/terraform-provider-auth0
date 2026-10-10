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

	if data.HasChange("status") {
		status := management.FeatureFlagStatusEnum(data.Get("status").(string))
		request.SetStatus(&status)
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
