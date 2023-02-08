package connection

import (
	"context"
	"reflect"
	"testing"
)

func TestConnectionInstanceStateUpgradeV0(t *testing.T) {
	for _, tt := range []struct {
		name            string
		version         interface{}
		versionExpected int
	}{
		{
			name:            "Empty",
			version:         "",
			versionExpected: 0,
		},
		{
			name:            "Zero",
			version:         "0",
			versionExpected: 0,
		},
		{
			name:            "NonZero",
			version:         "123",
			versionExpected: 123,
		},
		{
			name:            "Invalid",
			version:         "foo",
			versionExpected: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"strategy_version": tt.version},
				},
			}

			actual, err := connectionSchemaUpgradeV0(context.Background(), state, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			expected := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"strategy_version": tt.versionExpected},
				},
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
			}
		})
	}
}

func TestConnectionInstanceStateUpgradeV1(t *testing.T) {
	for _, tt := range []struct {
		name               string
		validation         map[string]string
		validationExpected []map[string][]interface{}
	}{
		{
			name: "Only Min",
			validation: map[string]string{
				"min": "5",
			},
			validationExpected: []map[string][]interface{}{
				{
					"username": []interface{}{
						map[string]string{
							"min": "5",
						},
					},
				},
			},
		},
		{
			name: "Min and Max",
			validation: map[string]string{
				"min": "5",
				"max": "10",
			},
			validationExpected: []map[string][]interface{}{
				{
					"username": []interface{}{
						map[string]string{
							"min": "5",
							"max": "10",
						},
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"validation": tt.validation},
				},
			}

			actual, err := connectionSchemaUpgradeV1(context.Background(), state, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			expected := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"validation": tt.validationExpected},
				},
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
			}
		})
	}
}
