package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/company/codegen-pro/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnalyzer(t *testing.T) {
	patterns := []string{"node_modules/", "*.log"}
	a := NewAnalyzer(patterns)

	assert.NotNil(t, a)
	assert.NotNil(t, a.parserRegistry)
	assert.Equal(t, patterns, a.ignorePatterns)
}

func TestAnalyzer_ShouldIgnore(t *testing.T) {
	patterns := []string{"node_modules/", "*.log", "dist/"}
	a := NewAnalyzer(patterns)

	tests := []struct {
		path     string
		expected bool
	}{
		{"node_modules/express/index.js", true},
		{".git/config", true},
		{"src/main.go", false},
		{"app.log", true},
		{"dist/bundle.js", true},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := a.shouldIgnore(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyzer_GetLanguageFromExtension(t *testing.T) {
	a := NewAnalyzer([]string{})

	tests := []struct {
		ext      string
		expected string
	}{
		{".go", "Go"},
		{".py", "Python"},
		{".js", "JavaScript"},
		{".ts", "TypeScript"},
		{".rs", "Rust"},
		{".java", "Java"},
		{".rb", "Ruby"},
		{".php", "PHP"},
		{".txt", ""},
		{".md", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := a.getLanguageFromExtension(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyzer_CountLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := `line 1
line 2
line 3
line 4
line 5`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	a := NewAnalyzer([]string{})
	lines, err := a.countLines(testFile)

	assert.NoError(t, err)
	assert.Equal(t, 5, lines)
}

func TestAnalyzer_ScanFiles(t *testing.T) {
	// Create test directory structure
	tmpDir := t.TempDir()

	// Create some files
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)

	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "src", "app.go"), []byte("package src"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "lib.js"), []byte("module.exports"), 0644)

	a := NewAnalyzer([]string{"node_modules/"})
	files, err := a.scanFiles(tmpDir)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 2) // Should find main.go and src/app.go

	// Should not contain node_modules files
	for _, file := range files {
		assert.NotContains(t, file, "node_modules")
	}
}

func TestAnalyzer_AnalyzeLanguages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n// comment\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "app.py"), []byte("print('hello')\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte("console.log('hi')\n"), 0644)

	files := []string{
		filepath.Join(tmpDir, "main.go"),
		filepath.Join(tmpDir, "app.py"),
		filepath.Join(tmpDir, "index.js"),
	}

	a := NewAnalyzer([]string{})
	result := &models.AnalysisResult{
		Languages: []models.LanguageInfo{},
	}

	err := a.analyzeLanguages(files, result)

	assert.NoError(t, err)
	assert.Len(t, result.Languages, 3)

	// Check that all languages are present
	langNames := make(map[string]bool)
	for _, lang := range result.Languages {
		langNames[lang.Name] = true
		assert.Greater(t, lang.FileCount, 0)
		assert.Greater(t, lang.LineCount, 0)
		assert.Greater(t, lang.Percentage, 0.0)
	}

	assert.True(t, langNames["Go"])
	assert.True(t, langNames["Python"])
	assert.True(t, langNames["JavaScript"])
}

func TestAnalyzer_DetectFramework(t *testing.T) {
	a := NewAnalyzer([]string{})

	tests := []struct {
		name         string
		dependencies []models.Dependency
		expected     string
	}{
		{
			name: "Express",
			dependencies: []models.Dependency{
				{Name: "express", Version: "4.18.0"},
			},
			expected: "Express",
		},
		{
			name: "Django",
			dependencies: []models.Dependency{
				{Name: "django", Version: "4.2.0"},
			},
			expected: "Django",
		},
		{
			name: "Gin",
			dependencies: []models.Dependency{
				{Name: "github.com/gin-gonic/gin", Version: "1.9.1"},
			},
			expected: "Gin",
		},
		{
			name:         "None",
			dependencies: []models.Dependency{},
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &models.AnalysisResult{
				Dependencies: tt.dependencies,
			}

			err := a.detectFramework([]string{}, result)
			assert.NoError(t, err)

			if tt.expected == "" {
				assert.Nil(t, result.Framework)
			} else {
				require.NotNil(t, result.Framework)
				assert.Equal(t, tt.expected, result.Framework.Name)
			}
		})
	}
}

func TestAnalyzer_FindEntryPoints(t *testing.T) {
	tmpDir := t.TempDir()

	// Create entry point files
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "app.py"), []byte("print('app')"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte("console.log('index')"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("text"), 0644)

	files := []string{
		filepath.Join(tmpDir, "main.go"),
		filepath.Join(tmpDir, "app.py"),
		filepath.Join(tmpDir, "index.js"),
		filepath.Join(tmpDir, "other.txt"),
	}

	a := NewAnalyzer([]string{})
	result := &models.AnalysisResult{
		EntryPoints: []models.EntryPoint{},
	}

	err := a.findEntryPoints(files, result)
	assert.NoError(t, err)

	// Should find at least main.go, app.py, index.js
	assert.GreaterOrEqual(t, len(result.EntryPoints), 3)

	// Check entry point properties
	for _, ep := range result.EntryPoints {
		assert.NotEmpty(t, ep.Path)
		assert.NotEmpty(t, ep.Language)
		assert.NotEmpty(t, ep.Type)
	}
}

func TestAnalyzer_ExtractConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config file with port
	configContent := `
server:
  port: 8080
  host: localhost
`
	os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0644)

	// Create main file with port
	mainContent := `
package main

const PORT = 3000
`
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainContent), 0644)

	files := []string{
		filepath.Join(tmpDir, "config.yaml"),
		filepath.Join(tmpDir, "main.go"),
	}

	a := NewAnalyzer([]string{})
	result := &models.AnalysisResult{
		Ports: []int{},
	}

	err := a.extractConfiguration(files, result)
	assert.NoError(t, err)

	// Should find at least one port
	assert.Greater(t, len(result.Ports), 0)

	// Check that ports are valid
	for _, port := range result.Ports {
		assert.Greater(t, port, 0)
		assert.Less(t, port, 65536)
	}
}

func TestAnalyzer_DeterminePrimaryLanguage(t *testing.T) {
	a := NewAnalyzer([]string{})

	result := &models.AnalysisResult{
		Languages: []models.LanguageInfo{
			{Name: "JavaScript", Percentage: 20.0},
			{Name: "Go", Percentage: 70.0},
			{Name: "Python", Percentage: 10.0},
		},
	}

	a.determinePrimaryLanguage(result)

	assert.Equal(t, "Go", result.PrimaryLanguage)
}

func TestAnalyzer_DetectBuildTools(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		files              []string
		expectedPkgManager string
		expectedBuildTool  string
	}{
		{
			name:               "Node.js",
			files:              []string{"package.json"},
			expectedPkgManager: "npm",
			expectedBuildTool:  "npm",
		},
		{
			name:               "Go",
			files:              []string{"go.mod"},
			expectedPkgManager: "go modules",
			expectedBuildTool:  "go",
		},
		{
			name:               "Python",
			files:              []string{"requirements.txt"},
			expectedPkgManager: "pip",
			expectedBuildTool:  "python",
		},
		{
			name:               "Rust",
			files:              []string{"Cargo.toml"},
			expectedPkgManager: "cargo",
			expectedBuildTool:  "cargo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create files
			var filePaths []string
			for _, file := range tt.files {
				path := filepath.Join(tmpDir, file)
				os.WriteFile(path, []byte("test"), 0644)
				filePaths = append(filePaths, path)
			}

			a := NewAnalyzer([]string{})
			result := &models.AnalysisResult{}

			a.detectBuildTools(filePaths, result)

			assert.Equal(t, tt.expectedPkgManager, result.PackageManager)
			assert.Equal(t, tt.expectedBuildTool, result.BuildTool)

			// Cleanup
			for _, path := range filePaths {
				os.Remove(path)
			}
		})
	}
}

func TestAnalyzer_Analyze_Integration(t *testing.T) {
	// Create a complete test project
	tmpDir := t.TempDir()

	// Create Go project structure
	os.MkdirAll(filepath.Join(tmpDir, "cmd", "app"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755)

	// Create go.mod
	goModContent := `module github.com/test/example

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)
`
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)

	// Create main.go
	mainContent := `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.Run(":8080")
}
`
	os.WriteFile(filepath.Join(tmpDir, "cmd", "app", "main.go"), []byte(mainContent), 0644)

	// Create another Go file
	os.WriteFile(filepath.Join(tmpDir, "internal", "handler.go"), []byte("package internal\n\nfunc Handle() {}"), 0644)

	// Run analysis
	a := NewAnalyzer([]string{})
	result, err := a.Analyze(tmpDir)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify results
	assert.Equal(t, tmpDir, result.ProjectPath)
	assert.NotEmpty(t, result.ProjectName)
	assert.Equal(t, "Go", result.PrimaryLanguage)

	// Should have detected Go language
	assert.Len(t, result.Languages, 1)
	assert.Equal(t, "Go", result.Languages[0].Name)

	// Should have parsed dependencies
	assert.Greater(t, len(result.Dependencies), 0)

	// Should have detected Gin framework
	if result.Framework != nil {
		assert.Equal(t, "Gin", result.Framework.Name)
	}

	// Should have found entry point
	assert.Greater(t, len(result.EntryPoints), 0)

	// Should have detected build tools
	assert.Equal(t, "go modules", result.PackageManager)
	assert.Equal(t, "go", result.BuildTool)

	// Should have file stats
	assert.Greater(t, result.FileStats.TotalFiles, 0)

	// Should have timestamp
	assert.False(t, result.AnalyzedAt.IsZero())
}
