package acctest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	var testCases = []struct {
		template string
		testName string
		expected string
	}{
		{
			template: "{{.testName}}",
			testName: "TestAccFoo",
			expected: "TestAccFoo",
		},
		{
			template: "{{.testName}",
			testName: "TestAccFoo",
			expected: "",
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("Test Case #%d", index), func(t *testing.T) {
			actual := ParseTestName(testCase.template, testCase.testName)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
