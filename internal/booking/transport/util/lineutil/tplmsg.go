package lineutil

import (
	"fmt"
	"html/template"
	"os"

	"github.com/goccy/go-yaml"
	"go.uber.org/zap/buffer"
)

var templateMap map[string]*template.Template

func LoadTemplateMsg(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}
	var templates map[string]string
	if err := yaml.Unmarshal(data, &templates); err != nil {
		return fmt.Errorf("failed to unmarshal templates: %w", err)
	}
	templateMap = make(map[string]*template.Template)
	for name, content := range templates {
		tmpl, err := template.New(name).Funcs(template.FuncMap{"add": add}).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		templateMap[name] = tmpl
	}
	return nil
}

func add(a, b int) int {
	return a + b
}

func RenderTemplate(name string, data interface{}) (string, error) {
	if templateMap == nil {
		return "", fmt.Errorf("template not loaded")
	}
	tmpl, ok := templateMap[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}
	var buf buffer.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	return buf.String(), nil
}
