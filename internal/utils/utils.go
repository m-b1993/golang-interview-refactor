package utils

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

func GetRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, "..", "..")
}

func GetConfigDir() string {
	rootDir := GetRootDir()
	return filepath.Join(rootDir, "config")
}

func GetTemplatesDir() string {
	rootDir := GetRootDir()
	return filepath.Join(rootDir, "static", "templates")
}

func RenderTemplate(pageData interface{}, templateName string) (string, error) {
	// Read and parse the HTML template file
	templatesDir := GetTemplatesDir()
	templatePath := filepath.Join(templatesDir, templateName)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Create a strings.Builder to store the rendered template
	var renderedTemplate strings.Builder

	err = tmpl.Execute(&renderedTemplate, pageData)
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Convert the rendered template to a string
	resultString := renderedTemplate.String()

	return resultString, nil
}
