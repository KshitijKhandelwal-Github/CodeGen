package generator

import (
	"strings"
	"text/template"
)

// getTemplateFuncs returns custom template functions for use in templates
func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// String manipulation
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
		"title":   strings.Title,

		// Math operations
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},

		// Utility functions
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
	}
}
