package common

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

// ExecuteTemplate renders a template with the given data
func ExecuteTemplate(tmpl *template.Template, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// ExecuteTemplateString renders a template string with the given data
func ExecuteTemplateString(templateStr string, data interface{}) ([]byte, error) {
	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return ExecuteTemplate(tmpl, data)
}

// FormatGoCode formats Go source code using gofmt
func FormatGoCode(code []byte) ([]byte, error) {
	formatted, err := format.Source(code)
	if err != nil {
		return nil, fmt.Errorf("failed to format code: %w", err)
	}
	return formatted, nil
}

// WriteGoFile writes Go source code to a file with formatting
func WriteGoFile(path string, content []byte) error {
	// Format the code first
	formatted, err := FormatGoCode(content)
	if err != nil {
		// If formatting fails, write unformatted code with a warning
		fmt.Printf("Warning: Failed to format %s: %v\n", path, err)
		formatted = content
	}

	// Write to file
	if err := os.WriteFile(path, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// AppendToGoFile appends Go source code to an existing file
func AppendToGoFile(path string, content []byte) error {
	// Read existing content
	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read existing file: %w", err)
	}

	// Append new content
	combined := append(existing, content...)

	// Format and write
	return WriteGoFile(path, combined)
}

// RenderAndFormatTemplate is a convenience function that renders a template and formats it
func RenderAndFormatTemplate(templateStr string, data interface{}) ([]byte, error) {
	// Execute template
	rendered, err := ExecuteTemplateString(templateStr, data)
	if err != nil {
		return nil, err
	}

	// Format code
	formatted, err := FormatGoCode(rendered)
	if err != nil {
		return nil, err
	}

	return formatted, nil
}

// WriteTemplateToFile renders a template and writes it to a file
func WriteTemplateToFile(templateStr string, data interface{}, outputPath string) error {
	formatted, err := RenderAndFormatTemplate(templateStr, data)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
