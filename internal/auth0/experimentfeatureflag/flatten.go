package experimentfeatureflag

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/auth0/go-auth0"
	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenFeatureFlag(data *schema.ResourceData, flag *management.GetFeatureFlagResponseContent) error {
	parameters, err := flattenParameters(flag.GetParameters())
	if err != nil {
		return err
	}

	result := multierror.Append(
		data.Set("name", flag.GetName()),
		data.Set("description", auth0.StringValue(flag.Description)),
		data.Set("type", string(flag.GetType())),
		data.Set("status", string(flag.GetStatus())),
		data.Set("parameters", parameters),
		data.Set("created_at", flag.GetCreatedAt().Format(time.RFC3339)),
		data.Set("updated_at", flag.GetUpdatedAt().Format(time.RFC3339)),
	)

	return result.ErrorOrNil()
}

func flattenParameters(parameters map[string]*management.FeatureFlagConfigParam) ([]interface{}, error) {
	result := make([]interface{}, 0, len(parameters))

	for name, cfg := range parameters {
		value, err := typedToString(cfg.GetValue())
		if err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"name":        name,
			"type":        string(cfg.GetType()),
			"value":       value,
			"description": cfg.GetDescription(),
		})
	}

	return result, nil
}

// typedToString renders a raw typed API value (bool, number, string, or object) as the string
// the Terraform schema stores. Numbers, objects, and arrays are rendered as compact JSON, so
// 1000000 stays "1000000" instead of the exponent form "1e+06".
func typedToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case bool:
		return strconv.FormatBool(v), nil
	case string:
		return v, nil
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(encoded), nil
	}
}

func flattenVariation(data *schema.ResourceData, variation *management.GetVariationResponseContent) error {
	overrides, err := flattenOverrides(variation.GetOverrides())
	if err != nil {
		return err
	}

	result := multierror.Append(
		data.Set("feature_flag_id", variation.GetFeatureFlagID()),
		data.Set("variation_id", variation.GetID()),
		data.Set("name", variation.GetName()),
		data.Set("description", auth0.StringValue(variation.Description)),
		data.Set("overrides", overrides),
		data.Set("created_at", variation.GetCreatedAt().Format(time.RFC3339)),
		data.Set("updated_at", variation.GetUpdatedAt().Format(time.RFC3339)),
	)

	return result.ErrorOrNil()
}

// flattenOverrides derives each override type from its JSON value, because the
// API returns only the value. Each JSON type maps to one parameter type.
func flattenOverrides(overrides map[string]interface{}) ([]interface{}, error) {
	result := make([]interface{}, 0, len(overrides))

	for name, raw := range overrides {
		var paramType management.FeatureFlagConfigParamTypeEnum
		switch raw.(type) {
		case bool:
			paramType = management.FeatureFlagConfigParamTypeEnumBoolean
		case float64:
			paramType = management.FeatureFlagConfigParamTypeEnumNumber
		case string:
			paramType = management.FeatureFlagConfigParamTypeEnumString
		case []interface{}:
			paramType = management.FeatureFlagConfigParamTypeEnumArray
		case map[string]interface{}:
			paramType = management.FeatureFlagConfigParamTypeEnumObject
		default:
			return nil, fmt.Errorf("override %q has an unsupported value type %T", name, raw)
		}

		value, err := typedToString(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"name":  name,
			"type":  string(paramType),
			"value": value,
		})
	}

	return result, nil
}
