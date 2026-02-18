package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	version = "0.1.0"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "codegen",
	Short: "CodeGen Pro - Automated project documentation and infrastructure generator",
	Long: `CodeGen Pro is a powerful CLI tool that analyzes your codebase and automatically generates:
	
  • Comprehensive README.md files
  • Production-ready Dockerfiles
  • Multi-cloud Terraform configurations (AWS, GCP, Azure)
  
It intelligently detects your programming language, framework, dependencies, 
and project structure to create accurate, customized outputs.`,
	Version: version,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(configCmd)
}
