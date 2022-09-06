package template

import "testing"

func TestTemplate(t *testing.T) {
	if s := ParseTestName(`{{.testName}}`, "foo"); s != "foo" {
		t.Errorf("unexpected result from template")
	}
}
