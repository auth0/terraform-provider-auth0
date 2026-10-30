package experiment

import (
	"encoding/json"
	"fmt"
	"time"

	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenExperiment(data *schema.ResourceData, experiment *management.GetExperimentResponseContent) error {
	allocations, err := flattenExperimentAllocations(experiment.GetAllocations())
	if err != nil {
		return err
	}

	flagSnapshot, err := flattenExperimentSnapshot(experiment.GetFeatureFlagSnapshot())
	if err != nil {
		return fmt.Errorf("feature_flag_snapshot: %w", err)
	}

	// Plugin SDK v2 stores a missing computed integer as zero, which is also a valid ramp level.
	result := multierror.Append(
		data.Set("name", experiment.GetName()),
		data.Set("description", experiment.GetDescription()),
		data.Set("feature_flag_id", experiment.GetFeatureFlagID()),
		data.Set("authentication_flow", experiment.GetAuthenticationFlow()),
		data.Set("default_config", string(experiment.GetDefaultConfig())),
		data.Set("allocation_strategy", string(experiment.GetAllocationStrategy())),
		data.Set("allocations", allocations),
		data.Set("levels", experiment.GetLevels()),
		data.Set("current_level", experiment.GetCurrentLevel()),
		data.Set("status", string(experiment.GetStatus())),
		data.Set("is_valid", experiment.GetIsValid()),
		data.Set("feature_flag_name", experiment.GetFeatureFlagName()),
		data.Set("feature_flag_snapshot", flagSnapshot),
		data.Set("editable_fields", experiment.GetEditableFields()),
		data.Set("started_at", experimentTimestamp(experiment.StartedAt)),
		data.Set("ended_at", experimentTimestamp(experiment.EndedAt)),
		data.Set("created_at", experiment.GetCreatedAt().Format(time.RFC3339)),
		data.Set("updated_at", experiment.GetUpdatedAt().Format(time.RFC3339)),
	)

	return result.ErrorOrNil()
}

func flattenExperimentAllocations(allocations []*management.AllocationItem) ([]interface{}, error) {
	result := make([]interface{}, 0, len(allocations))

	for index, allocation := range allocations {
		variationSnapshot, err := flattenExperimentSnapshot(allocation.GetVariationSnapshot())
		if err != nil {
			return nil, fmt.Errorf("allocations.%d.variation_snapshot: %w", index, err)
		}

		segmentSnapshot, err := flattenExperimentSnapshot(allocation.GetSegmentSnapshot())
		if err != nil {
			return nil, fmt.Errorf("allocations.%d.segment_snapshot: %w", index, err)
		}

		entry := map[string]interface{}{
			"variation_id":       allocation.GetVariationID(),
			"variation_name":     allocation.GetVariationName(),
			"segment_id":         allocation.GetSegmentID(),
			"segment_name":       allocation.GetSegmentName(),
			"is_control":         allocation.GetIsControl(),
			"is_fallback":        allocation.GetIsFallback(),
			"variation_snapshot": variationSnapshot,
			"segment_snapshot":   segmentSnapshot,
		}

		if allocation.Weight != nil {
			entry["weight"] = *allocation.Weight
		}
		if allocation.Priority != nil {
			entry["priority"] = *allocation.Priority
		}

		result = append(result, entry)
	}

	return result, nil
}

func flattenExperimentSnapshot(snapshot map[string]any) (string, error) {
	if snapshot == nil {
		return "", nil
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func experimentTimestamp(timestamp *time.Time) string {
	if timestamp == nil {
		return ""
	}

	return timestamp.Format(time.RFC3339)
}
