package template

import (
	"bytes"
	"text/template"
)

// ParseTestName renders templates defined with {{.testName}} placeholders.
// This is useful for acceptance tests.
func ParseTestName(rawTemplate, testName string) string {
	t := template.Must(template.New("tpl").Parse(rawTemplate))

	var buf bytes.Buffer
	err := t.Execute(&buf, map[string]string{"testName": testName})
	if err != nil {
		panic(err)
	}

	return buf.String()
}
