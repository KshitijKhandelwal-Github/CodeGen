package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysisResult_JSON(t *testing.T) {
	result := &AnalysisResult{
		ProjectPath:     "/path/to/project",
		ProjectName:     "test-project",
		PrimaryLanguage: "Go",
		Languages: []LanguageInfo{
			{
				Name:       "Go",
				Version:    "1.21",
				FileCount:  10,
				LineCount:  1000,
				Percentage: 100.0,
			},
		},
		Dependencies: []Dependency{
			{
				Name:     "github.com/spf13/cobra",
				Version:  "v1.8.0",
				Type:     "runtime",
				Registry: "Go modules",
			},
		},
		Ports:      []int{8080, 3000},
		AnalyzedAt: time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal back
	var decoded AnalysisResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.ProjectPath, decoded.ProjectPath)
	assert.Equal(t, result.ProjectName, decoded.ProjectName)
	assert.Equal(t, result.PrimaryLanguage, decoded.PrimaryLanguage)
	assert.Len(t, decoded.Languages, 1)
	assert.Len(t, decoded.Dependencies, 1)
}

func TestDependency_Types(t *testing.T) {
	tests := []struct {
		name    string
		dep     Dependency
		depType string
	}{
		{
			name: "Runtime dependency",
			dep: Dependency{
				Name:    "express",
				Version: "^4.0.0",
				Type:    "runtime",
			},
			depType: "runtime",
		},
		{
			name: "Dev dependency",
			dep: Dependency{
				Name:    "jest",
				Version: "^29.0.0",
				Type:    "dev",
			},
			depType: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.depType, tt.dep.Type)
		})
	}
}

func TestFrameworkInfo_JSON(t *testing.T) {
	framework := &FrameworkInfo{
		Name:     "Express",
		Version:  "4.18.0",
		Type:     "web",
		Features: []string{"routing", "middleware"},
	}

	data, err := json.Marshal(framework)
	require.NoError(t, err)

	var decoded FrameworkInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, framework.Name, decoded.Name)
	assert.Equal(t, framework.Type, decoded.Type)
}

func TestDatabaseInfo_NeverSerializeCredentials(t *testing.T) {
	db := &DatabaseInfo{
		Type: "postgresql",
		Host: "localhost",
		Port: 5432,
		Name: "mydb",
		Credentials: map[string]string{
    "username": "test_user",
    "password": "test_password_123",
},
	}

	// Marshal to JSON
	data, err := json.Marshal(db)
	require.NoError(t, err)

	// Credentials should NOT be in JSON
	assert.NotContains(t, string(data), "admin")
	assert.NotContains(t, string(data), "secret123")
	assert.NotContains(t, string(data), "credentials")

	// But other fields should be present
	assert.Contains(t, string(data), "postgresql")
	assert.Contains(t, string(data), "localhost")
}

func TestProjectConfig_Defaults(t *testing.T) {
	config := &ProjectConfig{
		ProjectName:    "test",
		GenerateREADME: true,
		Docker: DockerConfig{
			MultiStage:  true,
			HealthCheck: true,
		},
	}

	assert.Equal(t, "test", config.ProjectName)
	assert.True(t, config.GenerateREADME)
	assert.True(t, config.Docker.MultiStage)
}

func TestGenerationOptions(t *testing.T) {
	opts := &GenerationOptions{
		OutputDir: "/tmp",
		DryRun:    true,
		Overwrite: false,
		Verbose:   true,
	}

	assert.Equal(t, "/tmp", opts.OutputDir)
	assert.True(t, opts.DryRun)
	assert.False(t, opts.Overwrite)
	assert.True(t, opts.Verbose)
}

func TestGenerationResult(t *testing.T) {
	result := &GenerationResult{
		FilesCreated: []string{"README.md", "Dockerfile"},
		FilesSkipped: []string{"package.json"},
		Duration:     100 * time.Millisecond,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	assert.Len(t, result.FilesCreated, 2)
	assert.Len(t, result.FilesSkipped, 1)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.Equal(t, "1.0.0", result.Metadata["version"])
}
