package cmd

import (
	"fmt"
	"time"

	"github.com/company/codegen-pro/internal/analyzer"
	"github.com/company/codegen-pro/internal/config"
	"github.com/company/codegen-pro/internal/generator"
	"github.com/company/codegen-pro/internal/models"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	generateOutput    string
	generateDryRun    bool
	generateOverwrite bool
	generateREADME    bool
	generateDocker    bool
	generateTerraform bool
)

var generateCmd = &cobra.Command{
	Use:   "generate [path]",
	Short: "Generate documentation and infrastructure files",
	Long: `Analyzes the project and generates:
  • README.md with comprehensive documentation
  • Dockerfile with multi-stage builds
  • docker-compose.yml (if services detected)
  • Terraform configurations for AWS/GCP/Azure`,
	Args: cobra.MaximumNArgs(1),
	Run:  runGenerate,
}

func runGenerate(cmd *cobra.Command, args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	outputDir := generateOutput
	if outputDir == "" {
		outputDir = path
	}

	// Step 1: Load Configuration
	color.Cyan("Step 1/3: Loading configuration...")
	configMgr := config.NewManager()
	cfg, err := configMgr.LoadOrCreate(path)
	if err != nil {
		color.Red("✗ Error loading configuration: %v", err)
		return
	}

	// Override config with command-line flags if provided
	if cmd.Flags().Changed("readme") {
		cfg.GenerateREADME = generateREADME
	}
	if cmd.Flags().Changed("docker") {
		cfg.GenerateDocker = generateDocker
	}
	if cmd.Flags().Changed("terraform") {
		cfg.GenerateTerraform = generateTerraform
	}

	color.Green("✓ Configuration loaded")
	fmt.Println()

	// Step 2: Analyze Project
	color.Cyan("Step 2/3: Analyzing project...")

	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("  Scanning files..."),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(65*time.Millisecond),
	)

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

	a := analyzer.NewAnalyzer(cfg.IgnorePatterns)
	analysisResult, err := a.Analyze(path)
	done <- true

	if err != nil {
		color.Red("\n✗ Analysis failed: %v", err)
		return
	}

	color.Green("✓ Analysis complete")
	color.White("  Detected: %s project", analysisResult.PrimaryLanguage)
	if analysisResult.Framework != nil {
		color.White("  Framework: %s", analysisResult.Framework.Name)
	}
	fmt.Println()

	// Step 3: Generate Files
	color.Cyan("Step 3/3: Generating files...")

	gen, err := generator.NewGenerator()
	if err != nil {
		color.Red("✗ Error initializing generator: %v", err)
		return
	}

	opts := &models.GenerationOptions{
		OutputDir:      outputDir,
		DryRun:         generateDryRun,
		Overwrite:      generateOverwrite,
		Verbose:        verbose,
		Config:         cfg,
		AnalysisResult: analysisResult,
	}

	result, err := gen.Generate(opts)
	if err != nil {
		color.Red("✗ Generation failed: %v", err)
		return
	}

	// Display results
	fmt.Println()
	color.Cyan("═══════════════════════════════════════════════════")
	color.Cyan("           Generation Results")
	color.Cyan("═══════════════════════════════════════════════════")
	fmt.Println()

	if generateDryRun {
		color.Yellow("DRY RUN MODE - No files were actually created")
		fmt.Println()
	}

	if len(result.FilesCreated) > 0 {
		color.Green("Files Created:")
		for _, file := range result.FilesCreated {
			color.White("  ✓ %s", file)
		}
		fmt.Println()
	}

	if len(result.FilesSkipped) > 0 {
		color.Yellow("Files Skipped (already exist):")
		for _, file := range result.FilesSkipped {
			color.White("  - %s", file)
		}
		color.White("\nUse --overwrite to replace existing files")
		fmt.Println()
	}

	if len(result.Errors) > 0 {
		color.Red("Errors:")
		for _, err := range result.Errors {
			color.Red("  ✗ %s", err)
		}
		fmt.Println()
	}

	// Summary
	color.Cyan("═══════════════════════════════════════════════════")
	color.White("Duration: %v", result.Duration.Round(time.Millisecond))

	if len(result.Errors) == 0 {
		color.Green("\n✓ Generation complete!")
		if !generateDryRun {
			color.White("  Files are ready in: %s", outputDir)
		}
	} else {
		color.Yellow("\n⚠ Generation completed with errors")
	}
}

func init() {
	generateCmd.Flags().StringVarP(&generateOutput, "output", "o", "", "Output directory (default: current directory)")
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Preview what would be generated without creating files")
	generateCmd.Flags().BoolVar(&generateOverwrite, "overwrite", false, "Overwrite existing files")
	generateCmd.Flags().BoolVar(&generateREADME, "readme", true, "Generate README.md")
	generateCmd.Flags().BoolVar(&generateDocker, "docker", true, "Generate Docker files")
	generateCmd.Flags().BoolVar(&generateTerraform, "terraform", false, "Generate Terraform files")
}
