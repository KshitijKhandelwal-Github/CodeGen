package parsers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/company/codegen-pro/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageJSONParser_CanParse(t *testing.T) {
	parser := &PackageJSONParser{}

	assert.True(t, parser.CanParse("package.json"))
	assert.False(t, parser.CanParse("go.mod"))
	assert.False(t, parser.CanParse("requirements.txt"))
}

func TestPackageJSONParser_GetLanguage(t *testing.T) {
	parser := &PackageJSONParser{}
	assert.Equal(t, "JavaScript", parser.GetLanguage())
}

func TestPackageJSONParser_Parse(t *testing.T) {
	// Create test package.json
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")

	content := `{
		"name": "test",
		"dependencies": {
			"express": "^4.18.0",
			"lodash": "~4.17.21"
		},
		"devDependencies": {
			"jest": "^29.0.0"
		},
		"peerDependencies": {
			"react": ">=18.0.0"
		}
	}`

	err := os.WriteFile(pkgPath, []byte(content), 0644)
	require.NoError(t, err)

	parser := &PackageJSONParser{}
	deps, err := parser.Parse(pkgPath)

	assert.NoError(t, err)
	assert.Len(t, deps, 4)

	// Check runtime dependency
	var express *models.Dependency
	for i := range deps {
		if deps[i].Name == "express" {
			express = &deps[i]
			break
		}
	}
	require.NotNil(t, express)
	assert.Equal(t, "express", express.Name)
	assert.Equal(t, "^4.18.0", express.Version)
	assert.Equal(t, "runtime", express.Type)
	assert.Equal(t, "npm", express.Registry)

	// Check dev dependency
	var jest *models.Dependency
	for i := range deps {
		if deps[i].Name == "jest" {
			jest = &deps[i]
			break
		}
	}
	require.NotNil(t, jest)
	assert.Equal(t, "dev", jest.Type)
}

func TestRequirementsTxtParser_CanParse(t *testing.T) {
	parser := &RequirementsTxtParser{}

	assert.True(t, parser.CanParse("requirements.txt"))
	assert.True(t, parser.CanParse("requirements-dev.txt"))
	assert.False(t, parser.CanParse("package.json"))
}

func TestRequirementsTxtParser_Parse(t *testing.T) {
	tmpDir := t.TempDir()
	reqPath := filepath.Join(tmpDir, "requirements.txt")

	content := `flask==2.3.0
requests>=2.28.0
pytest~=7.3.0
# This is a comment
python-dotenv
-e git+https://github.com/user/repo.git#egg=package
`

	err := os.WriteFile(reqPath, []byte(content), 0644)
	require.NoError(t, err)

	parser := &RequirementsTxtParser{}
	deps, err := parser.Parse(reqPath)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(deps), 3)

	// Check flask
	var flask *models.Dependency
	for i := range deps {
		if deps[i].Name == "flask" {
			flask = &deps[i]
			break
		}
	}
	require.NotNil(t, flask)
	assert.Equal(t, "flask", flask.Name)
	assert.Equal(t, "==2.3.0", flask.Version)
	assert.Equal(t, "PyPI", flask.Registry)
}

func TestGoModParser_CanParse(t *testing.T) {
	parser := &GoModParser{}

	assert.True(t, parser.CanParse("go.mod"))
	assert.False(t, parser.CanParse("go.sum"))
	assert.False(t, parser.CanParse("package.json"))
}

func TestGoModParser_Parse(t *testing.T) {
	tmpDir := t.TempDir()
	modPath := filepath.Join(tmpDir, "go.mod")

	content := `module github.com/test/example

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/cobra v1.8.0
)

require (
	github.com/spf13/pflag v1.0.5 // indirect
)
`

	err := os.WriteFile(modPath, []byte(content), 0644)
	require.NoError(t, err)

	parser := &GoModParser{}
	deps, err := parser.Parse(modPath)

	assert.NoError(t, err)
	assert.Len(t, deps, 3)

	// Check direct dependency
	var cobra *models.Dependency
	for i := range deps {
		if deps[i].Name == "github.com/spf13/cobra" {
			cobra = &deps[i]
			break
		}
	}
	require.NotNil(t, cobra)
	assert.Equal(t, "github.com/spf13/cobra", cobra.Name)
	assert.Equal(t, "v1.8.0", cobra.Version)
	assert.Equal(t, "runtime", cobra.Type)

	// Check indirect dependency
	var pflag *models.Dependency
	for i := range deps {
		if deps[i].Name == "github.com/spf13/pflag" {
			pflag = &deps[i]
			break
		}
	}
	require.NotNil(t, pflag)
	assert.Equal(t, "dev", pflag.Type)
}

func TestCargoTomlParser_CanParse(t *testing.T) {
	parser := &CargoTomlParser{}

	assert.True(t, parser.CanParse("Cargo.toml"))
	assert.False(t, parser.CanParse("Cargo.lock"))
}

func TestParserRegistry_GetParser(t *testing.T) {
	registry := NewParserRegistry()

	tests := []struct {
		filename string
		expected string
	}{
		{"package.json", "JavaScript"},
		{"requirements.txt", "Python"},
		{"go.mod", "Go"},
		{"Cargo.toml", "Rust"},
		{"Gemfile", "Ruby"},
		{"composer.json", "PHP"},
		{"unknown.txt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			parser := registry.GetParser(tt.filename)
			if tt.expected == "" {
				assert.Nil(t, parser)
			} else {
				require.NotNil(t, parser)
				assert.Equal(t, tt.expected, parser.GetLanguage())
			}
		})
	}
}

func TestParserRegistry_ParseFile(t *testing.T) {
	registry := NewParserRegistry()

	// Test with real file
	tmpDir := t.TempDir()
	pkgPath := filepath.Join(tmpDir, "package.json")

	content := `{
		"dependencies": {
			"express": "^4.0.0"
		}
	}`

	err := os.WriteFile(pkgPath, []byte(content), 0644)
	require.NoError(t, err)

	deps, lang, err := registry.ParseFile(pkgPath)

	assert.NoError(t, err)
	assert.Equal(t, "JavaScript", lang)
	assert.Len(t, deps, 1)
	assert.Equal(t, "express", deps[0].Name)
}

func TestParserRegistry_ParseFile_UnknownFile(t *testing.T) {
	registry := NewParserRegistry()

	tmpDir := t.TempDir()
	unknownPath := filepath.Join(tmpDir, "unknown.txt")

	err := os.WriteFile(unknownPath, []byte("test"), 0644)
	require.NoError(t, err)

	_, _, err = registry.ParseFile(unknownPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no parser available")
}
