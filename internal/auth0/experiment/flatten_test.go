package experiment

import (
	"testing"
	"time"

	management "github.com/auth0/go-auth0/v3/management"
	"github.com/stretchr/testify/assert"
)

func TestFlattenExperimentAllocations(t *testing.T) {
	t.Run("it omits weight and priority when the API omits them", func(t *testing.T) {
		allocations, err := flattenExperimentAllocations([]*management.AllocationItem{
			{
				VariationID:   stringPointer("var_control"),
				VariationName: stringPointer("control"),
				IsControl:     boolPointer(true),
				IsFallback:    boolPointer(false),
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, []interface{}{
			map[string]interface{}{
				"variation_id":       "var_control",
				"variation_name":     "control",
				"segment_id":         "",
				"segment_name":       "",
				"is_control":         true,
				"is_fallback":        false,
				"variation_snapshot": "",
				"segment_snapshot":   "",
			},
		}, allocations)
	})

	t.Run("it sets weight, priority, and JSON-encoded snapshots", func(t *testing.T) {
		allocations, err := flattenExperimentAllocations([]*management.AllocationItem{
			{
				VariationID:       stringPointer("var_treatment"),
				VariationName:     stringPointer("treatment"),
				SegmentID:         stringPointer("seg_us"),
				SegmentName:       stringPointer("us-traffic"),
				Weight:            intPointer(30),
				Priority:          intPointer(1),
				IsControl:         boolPointer(false),
				IsFallback:        boolPointer(false),
				VariationSnapshot: map[string]any{"name": "treatment"},
				SegmentSnapshot:   map[string]any{"name": "us-traffic"},
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, []interface{}{
			map[string]interface{}{
				"variation_id":       "var_treatment",
				"variation_name":     "treatment",
				"segment_id":         "seg_us",
				"segment_name":       "us-traffic",
				"weight":             30,
				"priority":           1,
				"is_control":         false,
				"is_fallback":        false,
				"variation_snapshot": `{"name":"treatment"}`,
				"segment_snapshot":   `{"name":"us-traffic"}`,
			},
		}, allocations)
	})

	t.Run("it returns an empty list when there are no allocations", func(t *testing.T) {
		allocations, err := flattenExperimentAllocations(nil)

		assert.NoError(t, err)
		assert.Empty(t, allocations)
	})
}

func TestFlattenExperimentSnapshot(t *testing.T) {
	t.Run("it returns an empty string before the first activation", func(t *testing.T) {
		snapshot, err := flattenExperimentSnapshot(nil)

		assert.NoError(t, err)
		assert.Empty(t, snapshot)
	})

	t.Run("it encodes the snapshot as JSON", func(t *testing.T) {
		snapshot, err := flattenExperimentSnapshot(map[string]any{"name": "flag", "type": "self"})

		assert.NoError(t, err)
		assert.JSONEq(t, `{"name":"flag","type":"self"}`, snapshot)
	})
}

func TestExperimentTimestamp(t *testing.T) {
	t.Run("it returns an empty string when the API omits the timestamp", func(t *testing.T) {
		assert.Empty(t, experimentTimestamp(nil))
	})

	t.Run("it formats the timestamp as RFC 3339", func(t *testing.T) {
		timestamp := time.Date(2026, 10, 19, 12, 40, 57, 0, time.UTC)

		assert.Equal(t, "2026-10-19T12:40:57Z", experimentTimestamp(&timestamp))
	})
}
