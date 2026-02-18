package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/company/codegen-pro/internal/config"
	"github.com/spf13/cobra"
)

var configureAICmd = &cobra.Command{
	Use:   "ai",
	Short: "Configure AI settings",
	Long:  `Interactively configure AI enhancement settings including provider and API key.`,
	RunE:  runConfigureAI,
}

func init() {
	configCmd.AddCommand(configureAICmd)
}

func runConfigureAI(cmd *cobra.Command, args []string) error {
	projectPath := "."

	// Check if config exists
	if !config.ConfigExists(projectPath) {
		fmt.Println("❌ No configuration file found!")
		fmt.Println("   Run 'codegen init' first")
		return nil
	}

	// Load existing config
	mgr := config.NewManager()
	cfg, err := mgr.LoadOrCreate(projectPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║           AI Configuration Setup                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Current status
	if cfg.AI.Enabled {
		fmt.Printf("Current status: ✓ Enabled (%s)\n", cfg.AI.Provider)
		if cfg.AI.APIKey != "" {
			maskedKey := maskAPIKey(cfg.AI.APIKey)
			fmt.Printf("API Key: %s\n", maskedKey)
		} else {
			fmt.Println("API Key: (not set)")
		}
		fmt.Println()
	} else {
		fmt.Println("Current status: ✗ Disabled")
		fmt.Println()
	}

	// Enable/Disable AI
	currentlyEnabled := cfg.AI.Enabled
	enableAI := promptYesNo(reader, "🤖 Enable AI enhancement?", currentlyEnabled)
	cfg.AI.Enabled = enableAI

	if !enableAI {
		// Disable and save
		cfg.AI.APIKey = ""
		if err := mgr.Save(cfg, projectPath); err != nil {
			return fmt.Errorf("error saving config: %w", err)
		}
		fmt.Println()
		fmt.Println("✓ AI enhancement disabled")
		return nil
	}

	// Provider selection
	// Provider selection
fmt.Println()
fmt.Println("Choose AI Provider:")
fmt.Println("  1. Anthropic Claude (Recommended)")
fmt.Println("     - Best quality technical writing")
fmt.Println("     - Cost: ~$0.003 per README")
fmt.Println("     - Model: claude-sonnet-4-20250514")
fmt.Println()
fmt.Println("  2. OpenAI GPT-4")
fmt.Println("     - Good quality")
fmt.Println("     - Cost: ~$0.01 per README")
fmt.Println("     - Model: gpt-4")
fmt.Println()
fmt.Println("  3. Google Gemini (Cheapest)")
fmt.Println("     - Very good quality, fastest")
fmt.Println("     - Cost: ~$0.0001 per README (100x cheaper!)")
fmt.Println("     - Model: gemini-1.5-flash")
fmt.Println()

currentProvider := "1"
if cfg.AI.Provider == "openai" {
	currentProvider = "2"
} else if cfg.AI.Provider == "gemini" {
	currentProvider = "3"
}
fmt.Printf("Enter choice [%s]: ", currentProvider)
choice := strings.TrimSpace(readLine(reader))
if choice == "" {
	choice = currentProvider
}

switch choice {
case "2":
	cfg.AI.Provider = "openai"
	cfg.AI.Model = "gpt-4"
case "3":
	cfg.AI.Provider = "gemini"
	cfg.AI.Model = "gemini-1.5-flash-latest"
default:
	cfg.AI.Provider = "anthropic"
	cfg.AI.Model = "claude-sonnet-4-20250514"
}

// API Key
fmt.Println()
fmt.Printf("🔑 %s API Key:\n", strings.Title(cfg.AI.Provider))
fmt.Println()
fmt.Println("   Get your key from:")
switch cfg.AI.Provider {
case "anthropic":
	fmt.Println("   → https://console.anthropic.com/settings/keys")
	fmt.Println()
	fmt.Println("   Quick steps:")
	fmt.Println("   1. Sign up at console.anthropic.com")
	fmt.Println("   2. Go to Settings → API Keys")
	fmt.Println("   3. Create a new key")
	fmt.Println("   4. Copy and paste it below")
case "openai":
	fmt.Println("   → https://platform.openai.com/api-keys")
	fmt.Println()
	fmt.Println("   Quick steps:")
	fmt.Println("   1. Sign up at platform.openai.com")
	fmt.Println("   2. Go to API Keys")
	fmt.Println("   3. Create a new key")
	fmt.Println("   4. Copy and paste it below")
case "gemini":
	fmt.Println("   → https://aistudio.google.com/app/apikey")
	fmt.Println()
	fmt.Println("   Quick steps:")
	fmt.Println("   1. Go to Google AI Studio")
	fmt.Println("   2. Click 'Get API Key'")
	fmt.Println("   3. Create a new API key")
	fmt.Println("   4. Copy and paste it below")
}

	fmt.Println()
	if cfg.AI.APIKey != "" {
		maskedKey := maskAPIKey(cfg.AI.APIKey)
		fmt.Printf("   Current key: %s\n", maskedKey)
		fmt.Print("   Enter new key (or press Enter to keep current): ")
	} else {
		fmt.Print("   Enter key: ")
	}

	apiKey := strings.TrimSpace(readLine(reader))

	if apiKey != "" {
		// Validate key format
		if !isValidAPIKey(apiKey, cfg.AI.Provider) {
			fmt.Println()
			fmt.Println("⚠️  Warning: API key format doesn't look correct")
			if cfg.AI.Provider == "anthropic" {
				fmt.Println("   Anthropic keys should start with 'sk-ant-'")
			} else {
				fmt.Println("   OpenAI keys should start with 'sk-'")
			}
			fmt.Println()

			if !promptYesNo(reader, "   Continue anyway?", false) {
				fmt.Println("   Cancelled - no changes made")
				return nil
			}
		}

		cfg.AI.APIKey = apiKey
		fmt.Println()
		fmt.Println("   ✓ API key saved")
	} else if cfg.AI.APIKey == "" {
		fmt.Println()
		fmt.Println("   ⚠️  No API key set - AI enhancement will not work")
		fmt.Println("   You can add it later by running: codegen config ai")
	}

	// Save configuration
	if err := mgr.Save(cfg, projectPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ AI configuration saved")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: codegen generate --readme")
	fmt.Println("  2. Your README will include AI-generated descriptions!")
	fmt.Println()
	fmt.Println("💡 Security tip:")
	fmt.Println("   Add .codegen.yaml to your .gitignore")
	fmt.Println("   to keep your API key secure")
	fmt.Println()

	return nil
}

func isValidAPIKey(key, provider string) bool {
	key = strings.TrimSpace(key)
	switch provider {
	case "anthropic":
		return strings.HasPrefix(key, "sk-ant-") && len(key) > 20
	case "openai":
		return strings.HasPrefix(key, "sk-") && len(key) > 20
	case "gemini":
		return len(key) > 30 // Gemini keys are longer alphanumeric
	}
	return false
}
