package prompt_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/prompt"
)

func TestPromptScreenResourcesImportID(t *testing.T) {
	resources := map[string]*schema.Resource{
		"auth0_prompt_screen_partial":  prompt.NewScreenPartialResource(),
		"auth0_prompt_screen_renderer": prompt.NewPromptScreenRenderResource(),
	}

	var testCases = []struct {
		givenID     string
		expectError bool
	}{
		{givenID: "login-id:login-id"},
		{givenID: "login-id", expectError: true},
		{givenID: "login-id::login-id", expectError: true},
		{givenID: "login-id:", expectError: true},
		{givenID: ":login-id", expectError: true},
	}

	for resourceName, resource := range resources {
		for _, testCase := range testCases {
			t.Run(resourceName+"/"+testCase.givenID, func(t *testing.T) {
				data := schema.TestResourceDataRaw(t, resource.Schema, nil)
				data.SetId(testCase.givenID)

				actualData, err := resource.Importer.StateContext(context.Background(), data, nil)

				if testCase.expectError {
					assert.EqualError(
						t,
						err,
						`ID must be formatted as <prompt_type>:<screen_name>, got "`+testCase.givenID+`"`,
					)
					assert.Nil(t, actualData)
					return
				}

				assert.NoError(t, err)
				assert.Len(t, actualData, 1)
				assert.Equal(t, testCase.givenID, actualData[0].Id())
			})
		}
	}
}
