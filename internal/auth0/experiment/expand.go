package experiment

import (
	"fmt"

	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandExperiment(data *schema.ResourceData) (*management.CreateExperimentRequestContent, error) {
	config := data.GetRawConfig()
	allocations, err := expandExperimentAllocations(config.GetAttr("allocations"))
	if err != nil {
		return nil, err
	}

	request := &management.CreateExperimentRequestContent{
		Name:               data.Get("name").(string),
		FeatureFlagID:      data.Get("feature_flag_id").(string),
		Description:        value.String(config.GetAttr("description")),
		Allocations:        allocations,
		Levels:             expandExperimentLevels(data.Get("levels").([]interface{})),
		AuthenticationFlow: management.AuthenticationFlowEnum(data.Get("authentication_flow").(string)),
		AllocationStrategy: management.AllocationStrategyEnum(data.Get("allocation_strategy").(string)),
	}

	if defaultConfig := value.String(config.GetAttr("default_config")); defaultConfig != nil {
		configured := management.DefaultConfigEnum(*defaultConfig)
		request.DefaultConfig = &configured
	}

	return request, nil
}

// A nil result tells the caller there are no changed fields to patch.
func expandExperimentUpdate(data *schema.ResourceData) (*management.UpdateExperimentRequestParameters, error) {
	config := data.GetRawConfig()
	request := &management.UpdateExperimentRequestParameters{}

	if data.HasChange("name") {
		request.SetName(value.String(config.GetAttr("name")))
	}

	if data.HasChange("description") {
		request.SetDescription(value.String(config.GetAttr("description")))
	}

	if data.HasChange("authentication_flow") {
		flow := management.AuthenticationFlowEnum(data.Get("authentication_flow").(string))
		request.SetAuthenticationFlow(&flow)
	}

	if data.HasChange("default_config") {
		configured := value.String(config.GetAttr("default_config"))
		if configured == nil {
			return nil, fmt.Errorf("default_config cannot be removed from an existing experiment; " +
				"set it to another value or replace the experiment")
		}
		defaultConfig := management.DefaultConfigEnum(*configured)
		request.SetDefaultConfig(&defaultConfig)
	}

	if data.HasChange("allocations") {
		allocations, err := expandExperimentAllocations(config.GetAttr("allocations"))
		if err != nil {
			return nil, err
		}
		if len(allocations) == 0 {
			return nil, fmt.Errorf("allocations cannot be cleared with PATCH; removing them requires replacing the experiment")
		}
		request.SetAllocations(allocations)
	}

	if data.HasChange("levels") {
		levels := expandExperimentLevels(data.Get("levels").([]interface{}))
		if len(levels) == 0 {
			return nil, fmt.Errorf("levels cannot be removed from an existing experiment; " +
				"set other levels or replace the experiment")
		}
		request.SetLevels(levels)
	}

	if data.HasChange("status") {
		transition := management.ExperimentTransitionStatusEnum(data.Get("status").(string))
		request.SetStatus(&transition)
	}

	return request, nil
}

func expandExperimentAllocations(raw cty.Value) ([]*management.AllocationRequestItem, error) {
	if raw.IsNull() {
		return nil, nil
	}

	allocations := make([]*management.AllocationRequestItem, 0, raw.LengthInt())
	iterator := raw.ElementIterator()
	for iterator.Next() {
		_, entry := iterator.Element()
		allocations = append(allocations, &management.AllocationRequestItem{
			VariationID: entry.GetAttr("variation_id").AsString(),
			IsControl:   entry.GetAttr("is_control").True(),
			Weight:      value.Int(entry.GetAttr("weight")),
			SegmentID:   value.String(entry.GetAttr("segment_id")),
			Priority:    value.Int(entry.GetAttr("priority")),
			IsFallback:  value.Bool(entry.GetAttr("is_fallback")),
		})
	}

	return allocations, nil
}

func expandExperimentLevels(raw []interface{}) []int {
	if len(raw) == 0 {
		return nil
	}

	levels := make([]int, len(raw))
	for index, level := range raw {
		levels[index] = level.(int)
	}

	return levels
}
