package error

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/auth0/go-auth0/management"
	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/auth0/go-auth0/v3/management/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// newV3NotFoundError builds the error the v3 SDK returns on a 404, which wraps a
// *core.APIError instead of implementing the v1 management.Error interface.
func newV3NotFoundError(body string) error {
	return &managementv3.NotFoundError{
		APIError: core.NewAPIError(http.StatusNotFound, nil, errors.New(body)),
	}
}

var _ management.Error = &testManagementError{}

type testManagementError struct {
	StatusCode int
}

func (m testManagementError) Error() string {
	return fmt.Sprintf("%d", m.StatusCode)
}

func (m testManagementError) Status() int {
	return m.StatusCode
}

func TestHandleAPIError(t *testing.T) {
	testCases := []struct {
		name        string
		givenErr    error
		expectedErr error
	}{
		{
			name: "it returns nil if error is 404 and it triggers a resource deletion",
			givenErr: testManagementError{
				StatusCode: http.StatusNotFound,
			},
			expectedErr: nil,
		},
		{
			name: "it returns the error if error is 400 and it doesn't trigger a resource deletion",
			givenErr: testManagementError{
				StatusCode: http.StatusBadRequest,
			},
			expectedErr: fmt.Errorf("400"),
		},
		{
			name:        "it returns the error if the error is not a standard management error",
			givenErr:    fmt.Errorf("400"),
			expectedErr: fmt.Errorf("400"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, nil, nil)
			data.SetId("id")

			err := HandleAPIError(data, testCase.givenErr)

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
				assert.Equal(t, data.Id(), "id")
				return
			}

			assert.NoError(t, err)
			assert.Empty(t, data.Id())
		})
	}
}

func TestHandleReadAPIError(t *testing.T) {
	t.Run("it removes the resource from state and returns a warning if the error is a 404", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, nil, nil)
		data.SetId("id")

		diags := HandleReadAPIError("auth0_action", data, testManagementError{
			StatusCode: http.StatusNotFound,
		})

		assert.Empty(t, data.Id())
		assert.Len(t, diags, 1)
		assert.False(t, diags.HasError())
		assert.Equal(t, diag.Warning, diags[0].Severity)
		assert.Contains(t, diags[0].Detail, "auth0_action")
		assert.Contains(t, diags[0].Detail, "terraform state rm auth0_action.<name>")
	})

	t.Run("it returns the error and keeps the resource in state if the error is not a 404", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, nil, nil)
		data.SetId("id")

		diags := HandleReadAPIError("auth0_action", data, testManagementError{
			StatusCode: http.StatusBadRequest,
		})

		assert.Equal(t, "id", data.Id())
		assert.True(t, diags.HasError())
	})

	t.Run("it removes the resource from state and returns a warning on a v3 SDK 404", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, nil, nil)
		data.SetId("org_123::con_123")

		diags := HandleReadAPIError(
			"auth0_organization_connection",
			data,
			newV3NotFoundError(`{"statusCode":404,"error":"Not Found","message":"No connection found by that id"}`),
		)

		assert.Empty(t, data.Id())
		assert.Len(t, diags, 1)
		assert.False(t, diags.HasError())
		assert.Equal(t, diag.Warning, diags[0].Severity)
		assert.Contains(t, diags[0].Detail, "auth0_organization_connection")
		assert.Contains(t, diags[0].Detail, "org_123::con_123")
	})

	t.Run("it returns the error and keeps the resource in state on a non-404 v3 SDK error", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, nil, nil)
		data.SetId("id")

		diags := HandleReadAPIError("auth0_organization_connection", data, &managementv3.BadRequestError{
			APIError: core.NewAPIError(http.StatusBadRequest, nil, errors.New("bad request")),
		})

		assert.Equal(t, "id", data.Id())
		assert.True(t, diags.HasError())
	})
}

func TestRemoveFromStateWithWarning(t *testing.T) {
	data := schema.TestResourceDataRaw(t, nil, nil)
	data.SetId("con_123::client_456")

	diags := RemoveFromStateWithWarning(
		"auth0_connection_client",
		data,
		"the client is no longer enabled on the connection",
	)

	assert.Empty(t, data.Id())
	assert.Len(t, diags, 1)
	assert.False(t, diags.HasError())
	assert.Equal(t, diag.Warning, diags[0].Severity)
	assert.Contains(t, diags[0].Detail, "con_123::client_456")
	assert.Contains(t, diags[0].Detail, "the client is no longer enabled on the connection")
	assert.Contains(t, diags[0].Detail, "terraform state rm auth0_connection_client.<name>")
}

func TestIsStatusNotFound(t *testing.T) {
	testCases := []struct {
		name     string
		givenErr error
		expected bool
	}{
		{
			name:     "nil error",
			givenErr: nil,
			expected: false,
		},
		{
			name:     "v1 SDK 404",
			givenErr: testManagementError{StatusCode: http.StatusNotFound},
			expected: true,
		},
		{
			name:     "v1 SDK 400",
			givenErr: testManagementError{StatusCode: http.StatusBadRequest},
			expected: false,
		},
		{
			name:     "v3 SDK 404",
			givenErr: newV3NotFoundError(`{"statusCode":404,"error":"Not Found"}`),
			expected: true,
		},
		{
			name:     "v3 SDK bare core.APIError with a 404",
			givenErr: core.NewAPIError(http.StatusNotFound, nil, errors.New("not found")),
			expected: true,
		},
		{
			name:     "v3 SDK 404 wrapped by fmt.Errorf",
			givenErr: fmt.Errorf("reading connection: %w", newV3NotFoundError(`{"statusCode":404}`)),
			expected: true,
		},
		{
			name: "v3 SDK 400",
			givenErr: &managementv3.BadRequestError{
				APIError: core.NewAPIError(http.StatusBadRequest, nil, errors.New("bad request")),
			},
			expected: false,
		},
		{
			name:     "plain error",
			givenErr: errors.New("404"),
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, IsStatusNotFound(testCase.givenErr))
		})
	}
}
