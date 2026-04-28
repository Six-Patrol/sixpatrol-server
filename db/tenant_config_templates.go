package db

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"gorm.io/datatypes"
)

// RenderConfigTemplate walks a JSONMap and renders any string values that
// contain Go template placeholders (e.g. "{{ .ThresholdMs }}").
func RenderConfigTemplate(cfg datatypes.JSONMap, vars map[string]any) (datatypes.JSONMap, error) {
	rendered, err := renderValue(cfg, vars)
	if err != nil {
		return nil, err
	}

	result, ok := rendered.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected root config to be an object")
	}

	return datatypes.JSONMap(result), nil
}

func renderValue(value any, vars map[string]any) (any, error) {
	switch v := value.(type) {
	case datatypes.JSONMap:
		return renderMap(map[string]any(v), vars)
	case map[string]any:
		return renderMap(v, vars)
	case []any:
		return renderSlice(v, vars)
	case string:
		return renderTemplateString(v, vars)
	default:
		return v, nil
	}
}

func renderMap(input map[string]any, vars map[string]any) (map[string]any, error) {
	output := make(map[string]any, len(input))
	for key, value := range input {
		rendered, err := renderValue(value, vars)
		if err != nil {
			return nil, fmt.Errorf("render key %q: %w", key, err)
		}
		output[key] = rendered
	}
	return output, nil
}

func renderSlice(input []any, vars map[string]any) ([]any, error) {
	output := make([]any, len(input))
	for i, value := range input {
		rendered, err := renderValue(value, vars)
		if err != nil {
			return nil, fmt.Errorf("render index %d: %w", i, err)
		}
		output[i] = rendered
	}
	return output, nil
}

func renderTemplateString(input string, vars map[string]any) (string, error) {
	if !strings.Contains(input, "{{") {
		return input, nil
	}

	tmpl, err := template.New("cfg").Option("missingkey=error").Parse(input)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
