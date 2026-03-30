package generator

import (
	"os"
	"strings"
	"text/template"
)

func generateFile(filename, tmpl string, data TemplateData) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	template, err := template.New("endpoint").Parse(tmpl)
	if err != nil {
		return err
	}

	return template.Execute(f, data)
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(word[:1]))
			result.WriteString(strings.ToLower(word[1:]))
		}
	}
	return result.String()
}
