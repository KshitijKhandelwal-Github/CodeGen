package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/company/codegen-pro/internal/config"
	"github.com/company/codegen-pro/internal/models"
	"github.com/spf13/cobra"
)

var forceInit bool

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new CodeGen configuration file",
	Long: `Initialize a new CodeGen configuration file with interactive prompts.
You will be asked for project details and optional AI configuration.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing configuration")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Check if config already exists
	if !forceInit && config.ConfigExists(projectPath) {
		fmt.Println("⚠️  Configuration file already exists!")
		fmt.Println("   Use --force to overwrite, or edit .codegen.yaml manually")
		return nil
	}

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║          	CodeGen - Configuration Setup             ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create config with interactive prompts
	cfg, err := promptConfiguration(projectPath)
	if err != nil {
		return fmt.Errorf("configuration setup failed: %w", err)
	}

	// Save configuration
	mgr := config.NewManager()
	if err := mgr.Save(cfg, projectPath); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved to .codegen.yaml")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review/edit .codegen.yaml if needed")
	fmt.Println("  2. Run: codegen analyze")
	fmt.Println("  3. Run: codegen generate")
	fmt.Println()

	if cfg.AI.Enabled {
		fmt.Println("💡 Tip: Your API key is stored locally in .codegen.yaml")
		fmt.Println("   Add .codegen.yaml to .gitignore to keep it secure!")
		fmt.Println()
	}

	return nil
}

func promptConfiguration(projectPath string) (*models.ProjectConfig, error) {
	reader := bufio.NewReader(os.Stdin)

	// Infer project name from directory
	inferredName := getProjectNameFromPath(projectPath)

	cfg := &models.ProjectConfig{}

	// Project Name
	fmt.Printf("📦 Project name [%s]: ", inferredName)
	name := readLine(reader)
	if name == "" {
		name = inferredName
	}
	cfg.ProjectName = name

	// Description
	fmt.Print("📝 Description (optional): ")
	cfg.Description = readLine(reader)

	// Author
	fmt.Print("👤 Author (optional): ")
	cfg.Author = readLine(reader)

	// Version
	fmt.Print("🏷️  Version [0.1.0]: ")
	version := readLine(reader)
	if version == "" {
		version = "0.1.0"
	}
	cfg.Version = version

	// License
	fmt.Print("⚖️  License [MIT]: ")
	license := readLine(reader)
	if license == "" {
		license = "MIT"
	}
	cfg.License = license

	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Println("Generate Options:")
	fmt.Println()

	// Generate README
	cfg.GenerateREADME = promptYesNo(reader, "📄 Generate README.md?", true)

	// Generate Docker
	cfg.GenerateDocker = promptYesNo(reader, "🐳 Generate Docker files?", true)

	// Generate Terraform
	cfg.GenerateTerraform = promptYesNo(reader, "☁️  Generate Terraform configs?", false)

	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Println("AI Enhancement (Optional):")
	fmt.Println()

	// AI Configuration
	enableAI := promptYesNo(reader, "🤖 Enable AI-powered README enhancement?", false)
	cfg.AI.Enabled = enableAI

	// Update the AI provider selection in cmd/cli/cmd/init.go

// In the promptConfiguration function, update the provider selection:

if enableAI {
	// Provider
	fmt.Println()
	fmt.Println("Choose AI Provider:")
	fmt.Println("  1. Anthropic Claude (Recommended - $0.003/README)")
	fmt.Println("  2. OpenAI GPT-4 ($0.01/README)")
	fmt.Println("  3. Google Gemini (Cheapest - $0.0001/README)")
	fmt.Print("Enter choice [1]: ")
	choice := readLine(reader)
	
	switch choice {
	case "2":
		cfg.AI.Provider = "openai"
		cfg.AI.Model = "gpt-4"
	case "3":
		cfg.AI.Provider = "gemini"
		cfg.AI.Model = "gemini-1.5-flash"
	default:
		cfg.AI.Provider = "anthropic"
		cfg.AI.Model = "claude-sonnet-4-20250514"
	}

	// API Key
	fmt.Println()
	fmt.Printf("🔑 %s API Key: ", strings.Title(cfg.AI.Provider))
	fmt.Println("   Get your key from:")
	switch cfg.AI.Provider {
	case "anthropic":
		fmt.Println("   https://console.anthropic.com/settings/keys")
	case "openai":
		fmt.Println("   https://platform.openai.com/api-keys")
	case "gemini":
		fmt.Println("   https://aistudio.google.com/app/apikey")
	}
	fmt.Print("   Enter key (or press Enter to set later): ")
	
	apiKey := readLine(reader)
	if apiKey == "" {
		fmt.Println("   ⚠️  No API key provided - you can add it later in .codegen.yaml")
	} else {
		cfg.AI.APIKey = apiKey
		fmt.Println("   ✓ API key saved")
	}
}

	// Set defaults for other configs
	cfg.Docker = models.DockerConfig{
		MultiStage:  true,
		HealthCheck: true,
		ExposePort:  8080,
		WorkDir:     "/app",
	}

	cfg.Terraform = models.TerraformConfig{
		Provider:     "aws",
		Region:       "us-east-1",
		Environment:  "dev",
		CPU:          256,
		Memory:       512,
		MinInstances: 1,
		MaxInstances: 3,
	}

	cfg.README = models.READMEConfig{
		IncludeBadges:       true,
		IncludeTOC:          true,
		IncludeInstallation: true,
		IncludeUsage:        true,
		IncludeAPI:          true,
		IncludeContributing: false,
	}

	cfg.IgnorePatterns = []string{
		"node_modules/",
		"vendor/",
		".git/",
		"dist/",
		"build/",
		"__pycache__/",
		"*.pyc",
		".env",
		".env.local",
		"coverage/",
		".DS_Store",
	}

	return cfg, nil
}

func promptYesNo(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	var defaultStr string
	if defaultYes {
		defaultStr = "Y/n"
	} else {
		defaultStr = "y/N"
	}

	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	response := strings.ToLower(strings.TrimSpace(readLine(reader)))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func getProjectNameFromPath(path string) string {
	if path == "." || path == "" {
		// Get current directory name
		wd, err := os.Getwd()
		if err != nil {
			return "my-project"
		}
		return filepath.Base(wd)
	}
	return filepath.Base(path)
}
