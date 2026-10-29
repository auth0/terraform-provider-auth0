package agent

import (
	"encoding/json"
	"strings"
	"testing"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenMetadata(t *testing.T) {
	t.Run("it flattens an empty object to an empty string", func(t *testing.T) {
		got, err := flattenMetadata(managementv3.AgentMetadata{})
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("it flattens nil metadata to an empty string", func(t *testing.T) {
		got, err := flattenMetadata(nil)
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("it flattens metadata to JSON", func(t *testing.T) {
		got, err := flattenMetadata(managementv3.AgentMetadata{"team": "platform"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"team":"platform"}`, got)
	})
}

func TestExpandMetadata(t *testing.T) {
	empty := ""
	got, err := expandMetadata(&empty)
	require.NoError(t, err)
	assert.Nil(t, got)

	got, err = expandMetadata(nil)
	require.NoError(t, err)
	assert.Nil(t, got)

	raw := `{"team":"platform"}`
	got, err = expandMetadata(&raw)
	require.NoError(t, err)
	assert.Equal(t, managementv3.AgentMetadata{"team": "platform"}, got)
}

func TestValidateMetadata(t *testing.T) {
	tooLarge := `{"k":"` + strings.Repeat("a", maxMetadataBytes) + `"}`

	testCases := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{name: "valid object", input: `{"a":"b"}`},
		{name: "empty object", input: `{}`},
		{name: "invalid JSON", input: `{`, expectErr: true},
		{name: "array", input: `["a"]`, expectErr: true},
		{name: "string", input: `"a"`, expectErr: true},
		{name: "null", input: `null`, expectErr: true},
		{name: "too large", input: tooLarge, expectErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, errs := validateMetadata(testCase.input, "metadata")
			assert.Equal(t, testCase.expectErr, len(errs) > 0)
		})
	}
}

func TestValidateName(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{name: "valid", input: "My agent"},
		{name: "empty", input: "", expectErr: true},
		{name: "max length", input: strings.Repeat("a", maxNameLength)},
		{name: "too long", input: strings.Repeat("a", maxNameLength+1), expectErr: true},
		{name: "less than", input: "a<b", expectErr: true},
		{name: "greater than", input: "a>b", expectErr: true},
		{name: "null byte", input: "a\x00b", expectErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, errs := validateName(testCase.input, "name")
			assert.Equal(t, testCase.expectErr, len(errs) > 0)
		})
	}
}

func TestValidateExternalAgentID(t *testing.T) {
	_, errs := validateExternalAgentID("ext-1", "external_agent_id")
	assert.Empty(t, errs)

	_, errs = validateExternalAgentID("", "external_agent_id")
	assert.NotEmpty(t, errs)

	_, errs = validateExternalAgentID(strings.Repeat("a", maxExternalIDLen+1), "external_agent_id")
	assert.NotEmpty(t, errs)

	_, errs = validateExternalAgentID("a\x00b", "external_agent_id")
	assert.NotEmpty(t, errs)
}

// TestPatchSerializesExplicitNullMetadata pins that SetMetadata(nil) sends an explicit null,
// which is how removal of the metadata attribute clears it on the API.
func TestPatchSerializesExplicitNullMetadata(t *testing.T) {
	name := "agent"

	request := &managementv3.PatchAgentRequestParameters{Name: &name}
	b, err := json.Marshal(request)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"agent"}`, string(b))

	request.SetMetadata(nil)
	b, err = json.Marshal(request)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"agent","metadata":null}`, string(b))
}
