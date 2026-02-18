package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/company/codegen-pro/internal/models"
	"github.com/pelletier/go-toml/v2"
)

// DependencyParser is an interface for parsing dependency files
type DependencyParser interface {
	Parse(filePath string) ([]models.Dependency, error)
	CanParse(filename string) bool
	GetLanguage() string
}

// PackageJSONParser parses package.json for Node.js projects
type PackageJSONParser struct{}

func (p *PackageJSONParser) CanParse(filename string) bool {
	return filename == "package.json"
}

func (p *PackageJSONParser) GetLanguage() string {
	return "JavaScript"
}

func (p *PackageJSONParser) Parse(filePath string) ([]models.Dependency, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	var deps []models.Dependency

	// Regular dependencies
	for name, version := range pkg.Dependencies {
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "runtime",
			Registry: "npm",
		})
	}

	// Dev dependencies
	for name, version := range pkg.DevDependencies {
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "dev",
			Registry: "npm",
		})
	}

	// Peer dependencies
	for name, version := range pkg.PeerDependencies {
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "peer",
			Registry: "npm",
		})
	}

	// Optional dependencies
	for name, version := range pkg.OptionalDependencies {
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "optional",
			Registry: "npm",
		})
	}

	return deps, nil
}

// RequirementsTxtParser parses requirements.txt for Python projects
type RequirementsTxtParser struct{}

func (p *RequirementsTxtParser) CanParse(filename string) bool {
	return filename == "requirements.txt" || filename == "requirements-dev.txt"
}

func (p *RequirementsTxtParser) GetLanguage() string {
	return "Python"
}

func (p *RequirementsTxtParser) Parse(filePath string) ([]models.Dependency, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []models.Dependency
	scanner := bufio.NewScanner(file)

	// Regex to parse requirement lines: package_name==version or package_name>=version, etc.
	re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)(==|>=|<=|~=|!=)?(.+)?`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip -r or -e flags
		if strings.HasPrefix(line, "-r") || strings.HasPrefix(line, "-e") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			name := matches[1]
			version := ""
			if len(matches) >= 4 && matches[2] != "" {
				version = matches[2] + matches[3]
			}

			depType := "runtime"
			if strings.Contains(filepath.Base(filePath), "dev") {
				depType = "dev"
			}

			deps = append(deps, models.Dependency{
				Name:     name,
				Version:  version,
				Type:     depType,
				Registry: "PyPI",
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}

// PyprojectTomlParser parses pyproject.toml for Python projects
type PyprojectTomlParser struct{}

func (p *PyprojectTomlParser) CanParse(filename string) bool {
	return filename == "pyproject.toml"
}

func (p *PyprojectTomlParser) GetLanguage() string {
	return "Python"
}

func (p *PyprojectTomlParser) Parse(filePath string) ([]models.Dependency, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var pyproject struct {
		Project struct {
			Dependencies []string `toml:"dependencies"`
		} `toml:"project"`
		Tool struct {
			Poetry struct {
				Dependencies    map[string]interface{} `toml:"dependencies"`
				DevDependencies map[string]interface{} `toml:"dev-dependencies"`
			} `toml:"poetry"`
		} `toml:"tool"`
	}

	if err := toml.Unmarshal(data, &pyproject); err != nil {
		return nil, err
	}

	var deps []models.Dependency

	// Standard project dependencies
	for _, dep := range pyproject.Project.Dependencies {
		parts := strings.Fields(dep)
		if len(parts) > 0 {
			name := parts[0]
			version := ""
			if len(parts) > 1 {
				version = strings.Join(parts[1:], " ")
			}
			deps = append(deps, models.Dependency{
				Name:     name,
				Version:  version,
				Type:     "runtime",
				Registry: "PyPI",
			})
		}
	}

	// Poetry dependencies
	for name, versionInterface := range pyproject.Tool.Poetry.Dependencies {
		if name == "python" {
			continue // Skip python version constraint
		}

		version := ""
		switch v := versionInterface.(type) {
		case string:
			version = v
		case map[string]interface{}:
			if ver, ok := v["version"].(string); ok {
				version = ver
			}
		}

		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "runtime",
			Registry: "PyPI",
		})
	}

	// Poetry dev dependencies
	for name, versionInterface := range pyproject.Tool.Poetry.DevDependencies {
		version := ""
		switch v := versionInterface.(type) {
		case string:
			version = v
		case map[string]interface{}:
			if ver, ok := v["version"].(string); ok {
				version = ver
			}
		}

		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "dev",
			Registry: "PyPI",
		})
	}

	return deps, nil
}

// GoModParser parses go.mod for Go projects
type GoModParser struct{}

func (p *GoModParser) CanParse(filename string) bool {
	return filename == "go.mod"
}

func (p *GoModParser) GetLanguage() string {
	return "Go"
}

func (p *GoModParser) Parse(filePath string) ([]models.Dependency, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []models.Dependency
	scanner := bufio.NewScanner(file)
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Start of require block
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		// End of require block
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		// Parse require line
		if strings.HasPrefix(line, "require ") || inRequireBlock {
			// Remove "require " prefix if present
			line = strings.TrimPrefix(line, "require ")
			parts := strings.Fields(line)

			if len(parts) >= 2 {
				name := parts[0]
				version := parts[1]

				// Determine if it's a dev dependency (indirect)
				depType := "runtime"
				if strings.Contains(line, "// indirect") {
					depType = "dev"
				}

				deps = append(deps, models.Dependency{
					Name:     name,
					Version:  version,
					Type:     depType,
					Registry: "Go modules",
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}

// CargoTomlParser parses Cargo.toml for Rust projects
type CargoTomlParser struct{}

func (p *CargoTomlParser) CanParse(filename string) bool {
	return filename == "Cargo.toml"
}

func (p *CargoTomlParser) GetLanguage() string {
	return "Rust"
}

func (p *CargoTomlParser) Parse(filePath string) ([]models.Dependency, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cargo struct {
		Dependencies    map[string]interface{} `toml:"dependencies"`
		DevDependencies map[string]interface{} `toml:"dev-dependencies"`
	}

	if err := toml.Unmarshal(data, &cargo); err != nil {
		return nil, err
	}

	var deps []models.Dependency

	// Parse dependencies
	for name, versionInterface := range cargo.Dependencies {
		version := parseCargoVersion(versionInterface)
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "runtime",
			Registry: "crates.io",
		})
	}

	// Parse dev dependencies
	for name, versionInterface := range cargo.DevDependencies {
		version := parseCargoVersion(versionInterface)
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "dev",
			Registry: "crates.io",
		})
	}

	return deps, nil
}

// parseCargoVersion extracts version from various Cargo.toml formats
func parseCargoVersion(versionInterface interface{}) string {
	switch v := versionInterface.(type) {
	case string:
		return v
	case map[string]interface{}:
		if version, ok := v["version"].(string); ok {
			return version
		}
	}
	return ""
}

// GemfileParser parses Gemfile for Ruby projects
type GemfileParser struct{}

func (p *GemfileParser) CanParse(filename string) bool {
	return filename == "Gemfile"
}

func (p *GemfileParser) GetLanguage() string {
	return "Ruby"
}

func (p *GemfileParser) Parse(filePath string) ([]models.Dependency, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []models.Dependency
	scanner := bufio.NewScanner(file)

	// Regex to parse gem lines: gem 'name', 'version'
	re := regexp.MustCompile(`gem\s+['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]+)['"])?`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			name := matches[1]
			version := ""
			if len(matches) >= 3 {
				version = matches[2]
			}

			deps = append(deps, models.Dependency{
				Name:     name,
				Version:  version,
				Type:     "runtime",
				Registry: "RubyGems",
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return deps, nil
}

// ComposerJSONParser parses composer.json for PHP projects
type ComposerJSONParser struct{}

func (p *ComposerJSONParser) CanParse(filename string) bool {
	return filename == "composer.json"
}

func (p *ComposerJSONParser) GetLanguage() string {
	return "PHP"
}

func (p *ComposerJSONParser) Parse(filePath string) ([]models.Dependency, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var composer struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}

	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, err
	}

	var deps []models.Dependency

	// Runtime dependencies
	for name, version := range composer.Require {
		if name == "php" {
			continue // Skip PHP version
		}
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "runtime",
			Registry: "Packagist",
		})
	}

	// Dev dependencies
	for name, version := range composer.RequireDev {
		deps = append(deps, models.Dependency{
			Name:     name,
			Version:  version,
			Type:     "dev",
			Registry: "Packagist",
		})
	}

	return deps, nil
}

// ParserRegistry manages all dependency parsers
type ParserRegistry struct {
	parsers []DependencyParser
}

// NewParserRegistry creates a new parser registry with all parsers
func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: []DependencyParser{
			&PackageJSONParser{},
			&RequirementsTxtParser{},
			&PyprojectTomlParser{},
			&GoModParser{},
			&CargoTomlParser{},
			&GemfileParser{},
			&ComposerJSONParser{},
		},
	}
}

// GetParser returns the appropriate parser for a given filename
func (r *ParserRegistry) GetParser(filename string) DependencyParser {
	for _, parser := range r.parsers {
		if parser.CanParse(filename) {
			return parser
		}
	}
	return nil
}

// ParseFile parses a dependency file and returns dependencies
func (r *ParserRegistry) ParseFile(filePath string) ([]models.Dependency, string, error) {
	filename := filepath.Base(filePath)
	parser := r.GetParser(filename)

	if parser == nil {
		return nil, "", fmt.Errorf("no parser available for file: %s", filename)
	}

	deps, err := parser.Parse(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing %s: %w", filename, err)
	}

	return deps, parser.GetLanguage(), nil
}
