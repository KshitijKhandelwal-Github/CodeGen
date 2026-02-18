package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/company/codegen-pro/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()

	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr.viper)
}

func TestManager_Load_DefaultValues(t *testing.T) {
	mgr := NewManager()

	// Load without config file (should use defaults)
	config, err := mgr.Load()

	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Check default values
	assert.Equal(t, "my-project", config.ProjectName)
	assert.Equal(t, "0.1.0", config.Version)
	assert.Equal(t, "MIT", config.License)
	assert.True(t, config.GenerateREADME)
	assert.True(t, config.GenerateDocker)
	assert.False(t, config.GenerateTerraform)
}

func TestManager_Save(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager()

	config := &models.ProjectConfig{
		ProjectName:       "test-project",
		Version:           "1.0.0",
		License:           "Apache-2.0",
		Description:       "Test description",
		Author:            "Test Author",
		GenerateREADME:    true,
		GenerateDocker:    true,
		GenerateTerraform: true,
	}

	err := mgr.Save(config, tmpDir)
	assert.NoError(t, err)

	// Check that file was created
	configFile := filepath.Join(tmpDir, ".codegen.yaml")
	assert.FileExists(t, configFile)

	// Read it back
	mgr2 := NewManager()
	mgr2.viper.AddConfigPath(tmpDir)
	loaded, err := mgr2.Load()

	assert.NoError(t, err)
	assert.Equal(t, "test-project", loaded.ProjectName)
	assert.Equal(t, "1.0.0", loaded.Version)
	assert.Equal(t, "Apache-2.0", loaded.License)
}

func TestManager_LoadOrCreate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a project directory
	projectDir := filepath.Join(tmpDir, "my-awesome-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	mgr := NewManager()
	config, err := mgr.LoadOrCreate(projectDir)

	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Should have inferred project name from directory
	assert.Equal(t, "my-awesome-project", config.ProjectName)
}

func TestManager_CreateDefault(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager()

	config, err := mgr.CreateDefault(tmpDir)

	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Check that file was created
	configFile := filepath.Join(tmpDir, ".codegen.yaml")
	assert.FileExists(t, configFile)
}

func TestManager_GetSetMethods(t *testing.T) {
	mgr := NewManager()

	// Test Set and Get
	mgr.Set("test_key", "test_value")
	value := mgr.Get("test_key")
	assert.Equal(t, "test_value", value)

	// Test GetString
	mgr.Set("string_key", "string_value")
	strValue := mgr.GetString("string_key")
	assert.Equal(t, "string_value", strValue)

	// Test GetBool
	mgr.Set("bool_key", true)
	boolValue := mgr.GetBool("bool_key")
	assert.True(t, boolValue)

	// Test GetInt
	mgr.Set("int_key", 42)
	intValue := mgr.GetInt("int_key")
	assert.Equal(t, 42, intValue)
}

func TestConfigExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially should not exist
	assert.False(t, ConfigExists(tmpDir))

	// Create config file
	configFile := filepath.Join(tmpDir, ".codegen.yaml")
	err := os.WriteFile(configFile, []byte("project_name: test"), 0644)
	require.NoError(t, err)

	// Now should exist
	assert.True(t, ConfigExists(tmpDir))
}

func TestManager_DockerDefaults(t *testing.T) {
	mgr := NewManager()
	config, err := mgr.Load()

	assert.NoError(t, err)
	assert.True(t, config.Docker.MultiStage)
	assert.True(t, config.Docker.HealthCheck)
	assert.Equal(t, 8080, config.Docker.ExposePort)
	assert.Equal(t, "/app", config.Docker.WorkDir)
}

func TestManager_TerraformDefaults(t *testing.T) {
	mgr := NewManager()
	config, err := mgr.Load()

	assert.NoError(t, err)
	assert.Equal(t, "aws", config.Terraform.Provider)
	assert.Equal(t, "us-east-1", config.Terraform.Region)
	assert.Equal(t, "dev", config.Terraform.Environment)
	assert.Equal(t, 256, config.Terraform.CPU)
	assert.Equal(t, 512, config.Terraform.Memory)
	assert.Equal(t, 1, config.Terraform.MinInstances)
	assert.Equal(t, 3, config.Terraform.MaxInstances)
}

func TestManager_READMEDefaults(t *testing.T) {
	mgr := NewManager()
	config, err := mgr.Load()

	assert.NoError(t, err)
	assert.True(t, config.README.IncludeBadges)
	assert.True(t, config.README.IncludeTOC)
	assert.True(t, config.README.IncludeInstallation)
	assert.True(t, config.README.IncludeUsage)
	assert.True(t, config.README.IncludeAPI)
	assert.False(t, config.README.IncludeContributing)
}
