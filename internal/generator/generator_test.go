package generator

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/company/codegen-pro/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	gen, err := NewGenerator()

	assert.NoError(t, err)
	assert.NotNil(t, gen)
	assert.NotNil(t, gen.templates)
}

func TestGenerator_GenerateREADME(t *testing.T) {
	gen, err := NewGenerator()
	require.NoError(t, err)

	tmpDir := t.TempDir()

	opts := &models.GenerationOptions{
		OutputDir: tmpDir,
		DryRun:    false,
		Overwrite: false,
		Config: &models.ProjectConfig{
			ProjectName:    "test-project",
			Description:    "Test description",
			Author:         "Test Author",
			License:        "MIT",
			Version:        "1.0.0",
			GenerateREADME: true,
			README: models.READMEConfig{
				IncludeBadges:       true,
				IncludeTOC:          true,
				IncludeInstallation: true,
				IncludeUsage:        true,
			},
		},
		AnalysisResult: &models.AnalysisResult{
			ProjectName:     "test-project",
			PrimaryLanguage: "Go",
			Languages: []models.LanguageInfo{
				{Name: "Go", Percentage: 100.0, FileCount: 10, LineCount: 1000},
			},
			Dependencies: []models.Dependency{
				{Name: "github.com/spf13/cobra", Version: "v1.8.0"},
			},
			PackageManager: "go modules",
			BuildTool:      "go",
			Ports:          []int{8080},
		},
	}

	result := &models.GenerationResult{
		FilesCreated:  []string{},
		FilesSkipped:  []string{},
		FilesModified: []string{},
		Errors:        []string{},
	}

	err = gen.generateREADME(opts, result)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.FilesCreated), 1)
	if len(result.FilesCreated) > 0 {
		assert.Contains(t, result.FilesCreated[0], "README.md")
	}

	// Check that file exists
	readmePath := filepath.Join(tmpDir, "README.md")
	assert.FileExists(t, readmePath)

	// Check file content
	content, err := os.ReadFile(readmePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "test-project")
	assert.Contains(t, string(content), "Test description")
	assert.Contains(t, string(content), "MIT")
}

func TestGenerator_GenerateREADME_DryRun(t *testing.T) {
	gen, err := NewGenerator()
	require.NoError(t, err)

	tmpDir := t.TempDir()

	opts := &models.GenerationOptions{
		OutputDir: tmpDir,
		DryRun:    true, // Dry run mode
		Config: &models.ProjectConfig{
			ProjectName:    "test-project",
			GenerateREADME: true,
		},
		AnalysisResult: &models.AnalysisResult{
			ProjectName:     "test-project",
			PrimaryLanguage: "Go",
		},
	}

	result := &models.GenerationResult{
		FilesCreated: []string{},
	}

	err = gen.generateREADME(opts, result)

	assert.NoError(t, err)
	assert.Len(t, result.FilesCreated, 1)

	// File should NOT exist in dry run
	readmePath := filepath.Join(tmpDir, "README.md")
	assert.NoFileExists(t, readmePath)
}

func TestGenerator_GenerateREADME_SkipExisting(t *testing.T) {
	gen, err := NewGenerator()
	require.NoError(t, err)

	tmpDir := t.TempDir()

	// Create existing README
	readmePath := filepath.Join(tmpDir, "README.md")
	err = os.WriteFile(readmePath, []byte("existing content"), 0644)
	require.NoError(t, err)

	opts := &models.GenerationOptions{
		OutputDir: tmpDir,
		Overwrite: false, // Don't overwrite
		Config: &models.ProjectConfig{
			ProjectName:    "test-project",
			GenerateREADME: true,
		},
		AnalysisResult: &models.AnalysisResult{
			ProjectName: "test-project",
		},
	}

	result := &models.GenerationResult{
		FilesSkipped: []string{},
	}

	err = gen.generateREADME(opts, result)

	assert.NoError(t, err)
	assert.Len(t, result.FilesSkipped, 1)

	// Content should be unchanged
	content, _ := os.ReadFile(readmePath)
	assert.Equal(t, "existing content", string(content))
}

func TestGenerator_GenerateDocker(t *testing.T) {
	gen, err := NewGenerator()
	require.NoError(t, err)

	tmpDir := t.TempDir()

	opts := &models.GenerationOptions{
		OutputDir: tmpDir,
		Config: &models.ProjectConfig{
			ProjectName:    "test-project",
			GenerateDocker: true,
			Docker: models.DockerConfig{
				MultiStage:  true,
				HealthCheck: true,
				ExposePort:  8080,
				WorkDir:     "/app",
			},
		},
		AnalysisResult: &models.AnalysisResult{
			PrimaryLanguage: "Go",
			PackageManager:  "go modules",
			BuildTool:       "go",
			Ports:           []int{8080},
		},
	}

	result := &models.GenerationResult{
		FilesCreated: []string{},
	}

	err = gen.generateDocker(opts, result)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.FilesCreated), 2) // Dockerfile + .dockerignore

	// Check files exist
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	dockerignorePath := filepath.Join(tmpDir, ".dockerignore")

	assert.FileExists(t, dockerfilePath)
	assert.FileExists(t, dockerignorePath)

	// Check Dockerfile content
	content, _ := os.ReadFile(dockerfilePath)
	assert.Contains(t, string(content), "FROM")
	assert.Contains(t, string(content), "EXPOSE 8080")
}

func TestGenerator_Generate_Full(t *testing.T) {
	gen, err := NewGenerator()
	require.NoError(t, err)

	tmpDir := t.TempDir()

	opts := &models.GenerationOptions{
		OutputDir: tmpDir,
		Config: &models.ProjectConfig{
			ProjectName:       "full-test",
			GenerateREADME:    true,
			GenerateDocker:    true,
			GenerateTerraform: false,
		},
		AnalysisResult: &models.AnalysisResult{
			ProjectName:     "full-test",
			PrimaryLanguage: "Go",
		},
	}

	result, err := gen.Generate(opts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, len(result.FilesCreated), 0)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.NotEmpty(t, result.Metadata)
}

func TestGenerator_GetPort(t *testing.T) {
	gen, _ := NewGenerator()

	tests := []struct {
		name     string
		opts     *models.GenerationOptions
		expected int
	}{
		{
			name: "Config port",
			opts: &models.GenerationOptions{
				Config: &models.ProjectConfig{
					Docker: models.DockerConfig{ExposePort: 3000},
				},
				AnalysisResult: &models.AnalysisResult{Ports: []int{8080}},
			},
			expected: 3000,
		},
		{
			name: "Analysis port",
			opts: &models.GenerationOptions{
				Config: &models.ProjectConfig{
					Docker: models.DockerConfig{ExposePort: 0},
				},
				AnalysisResult: &models.AnalysisResult{Ports: []int{9000}},
			},
			expected: 9000,
		},
		{
			name: "Default port",
			opts: &models.GenerationOptions{
				Config: &models.ProjectConfig{
					Docker: models.DockerConfig{ExposePort: 0},
				},
				AnalysisResult: &models.AnalysisResult{Ports: []int{}},
			},
			expected: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := gen.getPort(tt.opts)
			assert.Equal(t, tt.expected, port)
		})
	}
}

func TestGenerator_FileExists(t *testing.T) {
	gen, _ := NewGenerator()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Initially doesn't exist
	assert.False(t, gen.fileExists(testFile))

	// Create file
	os.WriteFile(testFile, []byte("test"), 0644)

	// Now exists
	assert.True(t, gen.fileExists(testFile))
}
