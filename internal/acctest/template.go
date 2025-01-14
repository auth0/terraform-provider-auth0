package acctest

import (
	"bytes"
	"text/template"
)

// ParseTestName renders templates defined with {{.testName}} placeholders.
func ParseTestName(rawTemplate, testName string) string {
	t, err := template.New("tpl").Parse(rawTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{"testName": testName}); err != nil {
		return ""
	}

	return buf.String()
}

// ParseParametersInTemplate renders templates defined with placeholders present in paramDictionary as input.
func ParseParametersInTemplate(rawTemplate string, paramDictionary map[string]interface{}) string {
	t, err := template.New("tpl").Parse(rawTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, paramDictionary); err != nil {
		return ""
	}

	return buf.String()
}
