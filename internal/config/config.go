package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/company/codegen-pro/internal/models"
	"github.com/spf13/viper"
)

const (
	ConfigFileName = ".codegen"
	ConfigFileType = "yaml"
)

// Manager handles configuration loading and saving
type Manager struct {
	viper *viper.Viper
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	v := viper.New()
	v.SetConfigName(ConfigFileName)
	v.SetConfigType(ConfigFileType)
	v.AddConfigPath(".")     // Look in current directory
	v.AddConfigPath("$HOME") // Look in home directory

	// Set defaults
	setDefaults(v)

	return &Manager{viper: v}
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Project defaults
	v.SetDefault("project_name", "my-project")
	v.SetDefault("version", "0.1.0")
	v.SetDefault("license", "MIT")
	v.SetDefault("description", "")
	v.SetDefault("author", "")

	// Generation defaults
	v.SetDefault("generate_readme", true)
	v.SetDefault("generate_docker", true)
	v.SetDefault("generate_terraform", false)

	// Docker defaults
	v.SetDefault("docker.multi_stage", true)
	v.SetDefault("docker.health_check", true)
	v.SetDefault("docker.expose_port", 8080)
	v.SetDefault("docker.work_dir", "/app")

	// Terraform defaults
	v.SetDefault("terraform.provider", "aws")
	v.SetDefault("terraform.region", "us-east-1")
	v.SetDefault("terraform.environment", "dev")
	v.SetDefault("terraform.cpu", 256)
	v.SetDefault("terraform.memory", 512)
	v.SetDefault("terraform.min_instances", 1)
	v.SetDefault("terraform.max_instances", 3)

	// README defaults
	v.SetDefault("readme.include_badges", true)
	v.SetDefault("readme.include_toc", true)
	v.SetDefault("readme.include_installation", true)
	v.SetDefault("readme.include_usage", true)
	v.SetDefault("readme.include_api", true)
	v.SetDefault("readme.include_contributing", false)

	// AI defaults
	v.SetDefault("ai.enabled", false)
	v.SetDefault("ai.provider", "anthropic")
	v.SetDefault("ai.model", "claude-sonnet-4-20250514")
	v.SetDefault("ai.api_key", "")

	// Ignore patterns
	v.SetDefault("ignore_patterns", []string{
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
	})
}

// Load reads configuration from file
func (m *Manager) Load() (*models.ProjectConfig, error) {
	m.viper.AddConfigPath(".")
	m.viper.SetConfigName(".codegen")
	m.viper.SetConfigType("yaml")

	// Try to read config file
	if err := m.viper.ReadInConfig(); err != nil {
		fmt.Printf("DEBUG: ReadInConfig error: %v\n", err) // ← ADD THIS
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// ADD THIS DEBUG OUTPUT
	fmt.Printf("DEBUG: Config file used: %s\n", m.viper.ConfigFileUsed())
	fmt.Printf("DEBUG: generate_readme = %v\n", m.viper.Get("generate_readme"))
	fmt.Printf("DEBUG: generate_docker = %v\n", m.viper.Get("generate_docker"))
	fmt.Printf("DEBUG: project_name = %v\n", m.viper.Get("project_name"))

	var config models.ProjectConfig
	if err := m.viper.Unmarshal(&config); err != nil {
		fmt.Printf("DEBUG: Unmarshal error: %v\n", err) // ← ADD THIS
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// ADD THIS DEBUG OUTPUT
	fmt.Printf("DEBUG: After unmarshal - GenerateREADME = %v\n", config.GenerateREADME)
	fmt.Printf("DEBUG: After unmarshal - ProjectName = %v\n", config.ProjectName)

	return &config, nil
}

// LoadOrCreate loads existing config or creates default
func (m *Manager) LoadOrCreate(projectPath string) (*models.ProjectConfig, error) {
	// Add project path to search paths
	m.viper.AddConfigPath(projectPath)

	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	// If project name is still default, try to infer from directory
	if config.ProjectName == "my-project" || config.ProjectName == "" {
		dirName := filepath.Base(projectPath)
		if dirName != "." && dirName != "/" {
			config.ProjectName = dirName
		}
	}

	return config, nil
}

// Save writes configuration to file
func (m *Manager) Save(config *models.ProjectConfig, outputPath string) error {
	// CRITICAL: Use exact keys that match struct tags
	m.viper.Set("project_name", config.ProjectName)
	m.viper.Set("description", config.Description)
	m.viper.Set("author", config.Author)
	m.viper.Set("license", config.License)
	m.viper.Set("version", config.Version)

	// Generation options - EXACT key names
	m.viper.Set("generate_readme", config.GenerateREADME)
	m.viper.Set("generate_docker", config.GenerateDocker)
	m.viper.Set("generate_terraform", config.GenerateTerraform)

	// Docker config
	m.viper.Set("docker.multi_stage", config.Docker.MultiStage)
	m.viper.Set("docker.health_check", config.Docker.HealthCheck)
	m.viper.Set("docker.expose_port", config.Docker.ExposePort)
	m.viper.Set("docker.work_dir", config.Docker.WorkDir)

	// Terraform config
	m.viper.Set("terraform.provider", config.Terraform.Provider)
	m.viper.Set("terraform.region", config.Terraform.Region)
	m.viper.Set("terraform.environment", config.Terraform.Environment)
	m.viper.Set("terraform.cpu", config.Terraform.CPU)
	m.viper.Set("terraform.memory", config.Terraform.Memory)
	m.viper.Set("terraform.min_instances", config.Terraform.MinInstances)
	m.viper.Set("terraform.max_instances", config.Terraform.MaxInstances)

	// README config
	m.viper.Set("readme.include_badges", config.README.IncludeBadges)
	m.viper.Set("readme.include_toc", config.README.IncludeTOC)
	m.viper.Set("readme.include_installation", config.README.IncludeInstallation)
	m.viper.Set("readme.include_usage", config.README.IncludeUsage)
	m.viper.Set("readme.include_api", config.README.IncludeAPI)
	m.viper.Set("readme.include_contributing", config.README.IncludeContributing)

	// AI config
	m.viper.Set("ai.enabled", config.AI.Enabled)
	m.viper.Set("ai.provider", config.AI.Provider)
	m.viper.Set("ai.model", config.AI.Model)
	m.viper.Set("ai.api_key", config.AI.APIKey)

	// Ignore patterns
	m.viper.Set("ignore_patterns", config.IgnorePatterns)

	// Construct file path
	configFile := filepath.Join(outputPath, ConfigFileName+"."+ConfigFileType)

	// Write to file
	if err := m.viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// CreateDefault creates a default configuration file
func (m *Manager) CreateDefault(outputPath string) (*models.ProjectConfig, error) {
	// Get default config
	config, err := m.Load()
	if err != nil {
		return nil, err
	}

	// Try to infer project name from directory
	absPath, err := filepath.Abs(outputPath)
	if err == nil {
		dirName := filepath.Base(absPath)
		if dirName != "." && dirName != "/" {
			config.ProjectName = dirName
		}
	}

	// Save to file
	if err := m.Save(config, outputPath); err != nil {
		return nil, err
	}

	return config, nil
}

// Get retrieves a configuration value by key
func (m *Manager) Get(key string) interface{} {
	return m.viper.Get(key)
}

// GetString retrieves a string configuration value
func (m *Manager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetBool retrieves a boolean configuration value
func (m *Manager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// GetInt retrieves an integer configuration value
func (m *Manager) GetInt(key string) int {
	return m.viper.GetInt(key)
}

// Set sets a configuration value
func (m *Manager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
}

// ConfigExists checks if a config file exists in the given directory
func ConfigExists(dir string) bool {
	configPath := filepath.Join(dir, ConfigFileName+"."+ConfigFileType)
	_, err := os.Stat(configPath)
	return err == nil
}
