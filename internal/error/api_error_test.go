package error

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

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
}
