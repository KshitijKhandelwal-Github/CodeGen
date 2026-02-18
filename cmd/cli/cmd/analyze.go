package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/company/codegen-pro/internal/analyzer"
	"github.com/company/codegen-pro/internal/config"
	"github.com/company/codegen-pro/internal/models"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	analyzeOutput string
	analyzeJSON   bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Analyze a codebase and detect languages, frameworks, and dependencies",
	Long: `Analyzes a project directory to detect:
  • Programming languages and versions
  • Dependencies and package managers
  • Web frameworks and configurations
  • Entry points and build tools
  • Ports and environment variables
  • File statistics`,
	Args: cobra.MaximumNArgs(1),
	Run:  runAnalyze,
}

func runAnalyze(cmd *cobra.Command, args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Load configuration for ignore patterns
	configMgr := config.NewManager()
	cfg, err := configMgr.LoadOrCreate(path)
	if err != nil {
		color.Red("✗ Error loading configuration: %v", err)
		return
	}

	// Create analyzer
	a := analyzer.NewAnalyzer(cfg.IgnorePatterns)

	// Show progress
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Analyzing project..."),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(65*time.Millisecond),
	)

	// Start progress bar in goroutine
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				bar.Finish()
				return
			default:
				bar.Add(1)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Perform analysis
	result, err := a.Analyze(path)
	done <- true

	if err != nil {
		color.Red("\n✗ Analysis failed: %v", err)
		return
	}

	// Output results
	if analyzeJSON {
		outputJSON(result)
	} else {
		outputHuman(result)
	}

	// Save to file if requested
	if analyzeOutput != "" {
		if err := saveAnalysisToFile(result, analyzeOutput); err != nil {
			color.Red("\n✗ Error saving to file: %v", err)
		} else {
			color.Green("\n✓ Analysis saved to: %s", analyzeOutput)
		}
	}
}

func outputHuman(result *models.AnalysisResult) {
	fmt.Println()
	color.Cyan("═══════════════════════════════════════════════════")
	color.Cyan("           Analysis Results")
	color.Cyan("═══════════════════════════════════════════════════")
	fmt.Println()

	// Project Info
	color.White("Project: %s", result.ProjectName)
	color.White("Path:    %s", result.ProjectPath)
	fmt.Println()

	// Languages
	if len(result.Languages) > 0 {
		color.Yellow("Languages:")
		for _, lang := range result.Languages {
			color.White("  • %s: %.1f%% (%d files, %d lines)",
				lang.Name, lang.Percentage, lang.FileCount, lang.LineCount)
			if lang.Version != "" {
				color.White("    Version: %s", lang.Version)
			}
		}
		fmt.Println()
	}

	// Primary Language
	if result.PrimaryLanguage != "" {
		color.Green("Primary Language: %s", result.PrimaryLanguage)
		fmt.Println()
	}

	// Framework
	if result.Framework != nil {
		color.Yellow("Framework:")
		color.White("  Name:    %s %s", result.Framework.Name, result.Framework.Version)
		color.White("  Type:    %s", result.Framework.Type)
		if result.Framework.RoutesCount > 0 {
			color.White("  Routes:  %d", result.Framework.RoutesCount)
		}
		fmt.Println()
	}

	// Dependencies
	if len(result.Dependencies) > 0 {
		color.Yellow("Dependencies: %d total", len(result.Dependencies))

		// Group by type
		runtime := 0
		dev := 0
		for _, dep := range result.Dependencies {
			if dep.Type == "runtime" {
				runtime++
			} else if dep.Type == "dev" {
				dev++
			}
		}
		color.White("  Runtime: %d", runtime)
		color.White("  Dev:     %d", dev)
		fmt.Println()
	}

	// Build Tools
	if result.PackageManager != "" || result.BuildTool != "" {
		color.Yellow("Build Tools:")
		if result.PackageManager != "" {
			color.White("  Package Manager: %s", result.PackageManager)
		}
		if result.BuildTool != "" {
			color.White("  Build Tool:      %s", result.BuildTool)
		}
		fmt.Println()
	}

	// Entry Points
	if len(result.EntryPoints) > 0 {
		color.Yellow("Entry Points:")
		for _, ep := range result.EntryPoints {
			color.White("  • %s (%s)", ep.Path, ep.Language)
		}
		fmt.Println()
	}

	// Configuration
	if len(result.Ports) > 0 {
		color.Yellow("Detected Ports:")
		for _, port := range result.Ports {
			color.White("  • %d", port)
		}
		fmt.Println()
	}

	// Database
	if result.Database != nil {
		color.Yellow("Database:")
		color.White("  Type: %s", result.Database.Type)
		if result.Database.ORM != "" {
			color.White("  ORM:  %s", result.Database.ORM)
		}
		fmt.Println()
	}

	// File Statistics
	color.Yellow("File Statistics:")
	color.White("  Total Files:  %d", result.FileStats.TotalFiles)
	color.White("  Total Lines:  %d", result.FileStats.TotalLines)
	if result.FileStats.LargestFile != "" {
		color.White("  Largest File: %s (%d bytes)",
			result.FileStats.LargestFile, result.FileStats.LargestFileSize)
	}
	fmt.Println()

	color.Cyan("═══════════════════════════════════════════════════")
	color.Green("\n✓ Analysis complete!")
	color.White("  Use 'codegen generate' to create documentation and infrastructure files")
}

func outputJSON(result *models.AnalysisResult) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		color.Red("Error formatting JSON: %v", err)
		return
	}
	fmt.Println(string(data))
}

func saveAnalysisToFile(result *models.AnalysisResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "", "Save analysis to file")
	analyzeCmd.Flags().BoolVar(&analyzeJSON, "json", false, "Output in JSON format")
}
