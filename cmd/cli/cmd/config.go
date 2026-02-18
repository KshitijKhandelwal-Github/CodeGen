package cmd

import (
	"fmt"

	"github.com/company/codegen-pro/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and manage CodeGen Pro configuration.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show [path]",
	Short: "Show current configuration",
	Long:  `Display the current CodeGen Pro configuration.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	// configureAICmd is added in config_ai.go
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	mgr := config.NewManager()
	cfg, err := mgr.LoadOrCreate(projectPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("               Current Configuration")
	fmt.Println("═══════════════════════════════════════════════════════")

	fmt.Println("Project Settings:")
	fmt.Printf("  Name:        %s\n", cfg.ProjectName)
	fmt.Printf("  Version:     %s\n", cfg.Version)
	fmt.Printf("  License:     %s\n", cfg.License)
	if cfg.Description != "" {
		fmt.Printf("  Description: %s\n", cfg.Description)
	}
	if cfg.Author != "" {
		fmt.Printf("  Author:      %s\n", cfg.Author)
	}

	fmt.Println("\nGeneration Options:")
	fmt.Printf("  README:    %v\n", cfg.GenerateREADME)
	fmt.Printf("  Docker:    %v\n", cfg.GenerateDocker)
	fmt.Printf("  Terraform: %v\n", cfg.GenerateTerraform)

	fmt.Println("\nDocker Configuration:")
	fmt.Printf("  Port:        %d\n", cfg.Docker.ExposePort)
	fmt.Printf("  Multi-stage: %v\n", cfg.Docker.MultiStage)
	fmt.Printf("  Health Check: %v\n", cfg.Docker.HealthCheck)

	fmt.Println("\nAI Configuration:")
	fmt.Printf("  Enabled:  %v\n", cfg.AI.Enabled)
	if cfg.AI.Enabled {
		fmt.Printf("  Provider: %s\n", cfg.AI.Provider)
		fmt.Printf("  Model:    %s\n", cfg.AI.Model)

		// Mask API key for security
		if cfg.AI.APIKey != "" {
			masked := maskAPIKey(cfg.AI.APIKey)
			fmt.Printf("  API Key:  %s\n", masked)
		} else {
			fmt.Printf("  API Key:  (not set)\n")
		}
	}

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Edit .codegen.yaml to modify these settings")
	if cfg.AI.Enabled && cfg.AI.APIKey == "" {
		fmt.Println()
		fmt.Println("⚠️  AI is enabled but no API key is set")
		fmt.Println("   Run: codegen config ai")
	}
	fmt.Println()

	return nil
}

func maskAPIKey(key string) string {
	if len(key) < 12 {
		return "***"
	}
	return key[:7] + "..." + key[len(key)-4:]
}
