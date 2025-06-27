package core

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EmailTemplateData represents the data structure for email templates
type EmailTemplateData struct {
	Subject        string
	AppName        string
	Year           int
	RecipientEmail string
	User           interface{}
	LoginURL       string
	// Add more fields as needed for different email types
}

// EmailTemplateEngine handles email template rendering
type EmailTemplateEngine struct {
	templateDir string
	templates   map[string]*template.Template
}

// NewEmailTemplateEngine creates a new email template engine
func NewEmailTemplateEngine(templateDir string) *EmailTemplateEngine {
	return &EmailTemplateEngine{
		templateDir: templateDir,
		templates:   make(map[string]*template.Template),
	}
}

// Render renders an email template with the given data
func (e *EmailTemplateEngine) Render(templateName string, data EmailTemplateData) (string, error) {
	// Get or load template
	tmpl, err := e.getTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %v", templateName, err)
	}

	// Set default values
	if data.AppName == "" {
		data.AppName = "Base Laravel Go Project"
	}
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %v", templateName, err)
	}

	return buf.String(), nil
}

// getTemplate loads and caches a template
func (e *EmailTemplateEngine) getTemplate(templateName string) (*template.Template, error) {
	// Check if template is already cached
	if tmpl, exists := e.templates[templateName]; exists {
		return tmpl, nil
	}

	// Load base template
	baseTemplatePath := filepath.Join(e.templateDir, "base.html")
	baseTemplate, err := template.ParseFiles(baseTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %v", err)
	}

	// Load specific template
	specificTemplatePath := filepath.Join(e.templateDir, templateName+".html")
	if _, err := os.Stat(specificTemplatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template %s not found", templateName)
	}

	// Parse specific template into base template
	tmpl, err := baseTemplate.ParseFiles(specificTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %v", templateName, err)
	}

	// Cache the template
	e.templates[templateName] = tmpl

	return tmpl, nil
}

// PreloadTemplates preloads all templates in the directory
func (e *EmailTemplateEngine) PreloadTemplates() error {
	// Walk through the template directory
	err := filepath.Walk(e.templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, base template, and non-HTML files
		if info.IsDir() ||
			filepath.Base(path) == "base.html" ||
			!strings.HasSuffix(strings.ToLower(path), ".html") {
			return nil
		}

		// Get relative path from template directory
		relPath, err := filepath.Rel(e.templateDir, path)
		if err != nil {
			return err
		}

		// Remove .html extension to get template name
		templateName := relPath[:len(relPath)-5] // Remove .html

		// Load template
		_, err = e.getTemplate(templateName)
		if err != nil {
			return fmt.Errorf("failed to preload template %s: %v", templateName, err)
		}

		return nil
	})

	return err
}

// Global email template engine instance
var EmailTemplateEngineInstance *EmailTemplateEngine

// InitializeEmailTemplateEngine initializes the global email template engine
func InitializeEmailTemplateEngine() error {
	templateDir := "views/templates/mail"
	EmailTemplateEngineInstance = NewEmailTemplateEngine(templateDir)

	// Preload all templates
	return EmailTemplateEngineInstance.PreloadTemplates()
}

// RenderEmailTemplate is a convenience function to render email templates
func RenderEmailTemplate(templateName string, data EmailTemplateData) (string, error) {
	if EmailTemplateEngineInstance == nil {
		return "", fmt.Errorf("email template engine not initialized")
	}
	return EmailTemplateEngineInstance.Render(templateName, data)
}
