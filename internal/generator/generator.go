package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/company/codegen-pro/internal/ai" // ← ADD THIS
	"github.com/company/codegen-pro/internal/models"
)

//go:embed templates/*
var templatesFS embed.FS

// Generator handles file generation
type Generator struct {
	templates *template.Template
}

// NewGenerator creates a new generator instance
func NewGenerator() (*Generator, error) {
	// Parse templates with custom functions
	tmpl, err := template.New("").Funcs(getTemplateFuncs()).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error loading templates: %w", err)
	}

	return &Generator{
		templates: tmpl,
	}, nil
}

// Generate generates all requested files based on configuration
func (g *Generator) Generate(opts *models.GenerationOptions) (*models.GenerationResult, error) {
	startTime := time.Now()

	result := &models.GenerationResult{
		FilesCreated:  []string{},
		FilesSkipped:  []string{},
		FilesModified: []string{},
		Errors:        []string{},
		Metadata:      make(map[string]string),
	}

	// Create output directory if it doesn't exist
	if !opts.DryRun {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("error creating output directory: %w", err)
		}
	}

	// Generate README
	if opts.Config.GenerateREADME {
		if err := g.generateREADME(opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("README: %v", err))
		}
	}

	// Generate Docker files
	if opts.Config.GenerateDocker {
		if err := g.generateDocker(opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Docker: %v", err))
		}
	}

	// Generate Terraform files
	if opts.Config.GenerateTerraform {
		if err := g.generateTerraform(opts, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Terraform: %v", err))
		}
	}

	result.Duration = time.Since(startTime)
	result.Metadata["generator_version"] = "0.1.0"
	result.Metadata["generation_time"] = time.Now().Format(time.RFC3339)

	return result, nil
}

// generateREADME generates README.md file with optional AI enhancement
func (g *Generator) generateREADME(opts *models.GenerationOptions, result *models.GenerationResult) error {
	outputPath := filepath.Join(opts.OutputDir, "README.md")

	// Check if file exists and overwrite is not set
	if !opts.Overwrite && g.fileExists(outputPath) {
		result.FilesSkipped = append(result.FilesSkipped, "README.md")
		return nil
	}

	// Get AI-enhanced description if enabled
	aiDescription := ""
	if opts.Config.AI.Enabled {
		// Pass the user's API key from config
		aiClient, err := ai.NewClient(
			opts.Config.AI.Provider,
			opts.Config.AI.Model,
			opts.Config.AI.APIKey, // User's API key from .codegen.yaml
		)
		if err != nil {
			// If AI fails, continue without it but log the error
			if opts.Verbose {
				fmt.Printf("⚠️  AI enhancement unavailable: %v\n", err)
				fmt.Printf("    Set your API key in .codegen.yaml under ai.api_key\n")
				fmt.Printf("    Or use environment variable: ANTHROPIC_API_KEY or OPENAI_API_KEY\n")
			}
		} else {
			if opts.Verbose {
				fmt.Println("🤖 Calling AI to enhance README description...")
			}
			desc, err := aiClient.EnhanceREADME(opts.AnalysisResult, opts.Config)
			if err != nil {
				if opts.Verbose {
					fmt.Printf("⚠️  AI enhancement failed: %v\n", err)
					fmt.Println("    Continuing with standard template...")
				}
			} else {
				aiDescription = desc
				if opts.Verbose {
					fmt.Println("✓ AI enhancement successful")
				}
			}
		}
	}

	// Prepare template data
	data := map[string]interface{}{
		"ProjectName":    opts.Config.ProjectName,
		"Description":    opts.Config.Description,
		"AIDescription":  aiDescription, // AI-generated comprehensive description
		"Author":         opts.Config.Author,
		"License":        opts.Config.License,
		"Version":        opts.Config.Version,
		"Analysis":       opts.AnalysisResult,
		"Config":         opts.Config.README,
		"PrimaryLang":    opts.AnalysisResult.PrimaryLanguage,
		"Dependencies":   opts.AnalysisResult.Dependencies,
		"Framework":      opts.AnalysisResult.Framework,
		"EntryPoints":    opts.AnalysisResult.EntryPoints,
		"BuildTool":      opts.AnalysisResult.BuildTool,
		"PackageManager": opts.AnalysisResult.PackageManager,
	}

	if opts.DryRun {
		result.FilesCreated = append(result.FilesCreated, "README.md (dry-run)")
		return nil
	}

	// Execute template
	if err := g.executeTemplate("readme.tmpl", data, outputPath); err != nil {
		return err
	}

	result.FilesCreated = append(result.FilesCreated, "README.md")
	return nil
}

// generateDocker generates Dockerfile and .dockerignore
func (g *Generator) generateDocker(opts *models.GenerationOptions, result *models.GenerationResult) error {
	// Generate Dockerfile
	dockerfilePath := filepath.Join(opts.OutputDir, "Dockerfile")

	if !opts.Overwrite && g.fileExists(dockerfilePath) {
		result.FilesSkipped = append(result.FilesSkipped, "Dockerfile")
	} else {
		data := map[string]interface{}{
			"PrimaryLang":    opts.AnalysisResult.PrimaryLanguage,
			"Port":           g.getPort(opts),
			"MultiStage":     opts.Config.Docker.MultiStage,
			"HealthCheck":    opts.Config.Docker.HealthCheck,
			"WorkDir":        opts.Config.Docker.WorkDir,
			"PackageManager": opts.AnalysisResult.PackageManager,
		}

		if !opts.DryRun {
			if err := g.executeTemplate("dockerfile.tmpl", data, dockerfilePath); err != nil {
				return err
			}
		}
		result.FilesCreated = append(result.FilesCreated, "Dockerfile")
	}

	// Generate .dockerignore
	dockerignorePath := filepath.Join(opts.OutputDir, ".dockerignore")

	if !opts.Overwrite && g.fileExists(dockerignorePath) {
		result.FilesSkipped = append(result.FilesSkipped, ".dockerignore")
	} else {
		data := map[string]interface{}{
			"PrimaryLang": opts.AnalysisResult.PrimaryLanguage,
		}

		if !opts.DryRun {
			if err := g.executeTemplate("dockerignore.tmpl", data, dockerignorePath); err != nil {
				return err
			}
		}
		result.FilesCreated = append(result.FilesCreated, ".dockerignore")
	}

	return nil
}

// generateTerraform generates Terraform configuration files
func (g *Generator) generateTerraform(opts *models.GenerationOptions, result *models.GenerationResult) error {
	// Create terraform directory
	terraformDir := filepath.Join(opts.OutputDir, "terraform")
	if !opts.DryRun {
		if err := os.MkdirAll(terraformDir, 0755); err != nil {
			return fmt.Errorf("error creating terraform directory: %w", err)
		}
	}

	data := map[string]interface{}{
		"ProjectName": opts.Config.ProjectName,
		"Port":        g.getPort(opts),
		"Config":      opts.Config.Terraform,
	}

	// Generate main.tf
	mainPath := filepath.Join(terraformDir, "main.tf")
	if !opts.DryRun {
		if err := g.executeTemplate("terraform-aws-main.tmpl", data, mainPath); err != nil {
			return err
		}
	}
	result.FilesCreated = append(result.FilesCreated, "terraform/main.tf")

	// Generate outputs.tf
	outputsPath := filepath.Join(terraformDir, "outputs.tf")
	if !opts.DryRun {
		if err := g.executeTemplate("terraform-aws-outputs.tmpl", data, outputsPath); err != nil {
			return err
		}
	}
	result.FilesCreated = append(result.FilesCreated, "terraform/outputs.tf")

	// Generate variables.tf
	variablesPath := filepath.Join(terraformDir, "variables.tf")
	if !opts.DryRun {
		if err := g.executeTemplate("terraform-variables.tmpl", data, variablesPath); err != nil {
			return err
		}
	}
	result.FilesCreated = append(result.FilesCreated, "terraform/variables.tf")

	return nil
}

// executeTemplate executes a template and writes to a file
func (g *Generator) executeTemplate(templateName string, data interface{}, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", outputPath, err)
	}
	defer file.Close()

	if err := g.templates.ExecuteTemplate(file, templateName, data); err != nil {
		return fmt.Errorf("error executing template %s: %w", templateName, err)
	}

	return nil
}

// fileExists checks if a file exists
func (g *Generator) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getPort gets the port to use (priority: config > analysis > default)
func (g *Generator) getPort(opts *models.GenerationOptions) int {
	if opts.Config.Docker.ExposePort > 0 {
		return opts.Config.Docker.ExposePort
	}
	if len(opts.AnalysisResult.Ports) > 0 {
		return opts.AnalysisResult.Ports[0]
	}
	return 8080 // default
}
