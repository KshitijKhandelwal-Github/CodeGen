package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/company/codegen-pro/internal/models"
	"github.com/company/codegen-pro/pkg/parsers"
)

// Analyzer performs comprehensive codebase analysis
type Analyzer struct {
	parserRegistry *parsers.ParserRegistry
	ignorePatterns []string
	mu             sync.Mutex
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(ignorePatterns []string) *Analyzer {
	return &Analyzer{
		parserRegistry: parsers.NewParserRegistry(),
		ignorePatterns: ignorePatterns,
	}
}

// Analyze performs complete analysis of a project
func (a *Analyzer) Analyze(projectPath string) (*models.AnalysisResult, error) {
	result := &models.AnalysisResult{
		ProjectPath:   projectPath,
		ProjectName:   filepath.Base(projectPath),
		Languages:     []models.LanguageInfo{},
		Dependencies:  []models.Dependency{},
		EntryPoints:   []models.EntryPoint{},
		Configuration: make(map[string]interface{}),
		Ports:         []int{},
		Environment:   make(map[string]string),
		FileStats: models.FileStatistics{
			FilesByLanguage: make(map[string]int),
		},
		AnalyzedAt: time.Now(),
	}

	// Scan all files
	files, err := a.scanFiles(projectPath)
	if err != nil {
		return nil, fmt.Errorf("error scanning files: %w", err)
	}

	// Analyze languages
	if err := a.analyzeLanguages(files, result); err != nil {
		return nil, fmt.Errorf("error analyzing languages: %w", err)
	}

	// Parse dependencies
	if err := a.parseDependencies(files, result); err != nil {
		return nil, fmt.Errorf("error parsing dependencies: %w", err)
	}

	// Detect framework
	if err := a.detectFramework(files, result); err != nil {
		return nil, fmt.Errorf("error detecting framework: %w", err)
	}

	// Find entry points
	if err := a.findEntryPoints(files, result); err != nil {
		return nil, fmt.Errorf("error finding entry points: %w", err)
	}

	// Extract configuration
	if err := a.extractConfiguration(files, result); err != nil {
		return nil, fmt.Errorf("error extracting configuration: %w", err)
	}

	// Calculate file statistics
	if err := a.calculateFileStats(files, result); err != nil {
		return nil, fmt.Errorf("error calculating file stats: %w", err)
	}

	// Determine primary language
	a.determinePrimaryLanguage(result)

	// Detect package manager and build tool
	a.detectBuildTools(files, result)

	return result, nil
}

// scanFiles recursively scans all files in the project
func (a *Analyzer) scanFiles(rootPath string) ([]string, error) {
	var files []string
	var mu sync.Mutex

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Check if directory should be ignored
			relPath, _ := filepath.Rel(rootPath, path)
			if a.shouldIgnore(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be ignored
		relPath, _ := filepath.Rel(rootPath, path)
		if a.shouldIgnore(relPath) {
			return nil
		}

		mu.Lock()
		files = append(files, path)
		mu.Unlock()

		return nil
	})

	return files, err
}

// shouldIgnore checks if a path matches ignore patterns
func (a *Analyzer) shouldIgnore(path string) bool {
	// Always ignore .git
	if strings.Contains(path, ".git") {
		return true
	}

	for _, pattern := range a.ignorePatterns {
		// Simple pattern matching (can be enhanced with glob)
		if strings.HasSuffix(pattern, "/") {
			// Directory pattern
			if strings.HasPrefix(path, strings.TrimSuffix(pattern, "/")) {
				return true
			}
		} else {
			// File pattern
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				return true
			}
		}
	}

	return false
}

// analyzeLanguages detects programming languages used
func (a *Analyzer) analyzeLanguages(files []string, result *models.AnalysisResult) error {
	langStats := make(map[string]*models.LanguageInfo)

	for _, file := range files {
		ext := filepath.Ext(file)
		lang := a.getLanguageFromExtension(ext)

		if lang == "" {
			continue
		}

		if _, exists := langStats[lang]; !exists {
			langStats[lang] = &models.LanguageInfo{
				Name:       lang,
				Extensions: []string{ext},
			}
		}

		langStats[lang].FileCount++

		// Count lines (for statistics)
		lines, err := a.countLines(file)
		if err == nil {
			langStats[lang].LineCount += lines
		}
	}

	// Convert to slice and calculate percentages
	totalLines := 0
	for _, info := range langStats {
		totalLines += info.LineCount
	}

	for _, info := range langStats {
		if totalLines > 0 {
			info.Percentage = float64(info.LineCount) / float64(totalLines) * 100
		}
		result.Languages = append(result.Languages, *info)
	}

	return nil
}

// getLanguageFromExtension maps file extensions to languages
func (a *Analyzer) getLanguageFromExtension(ext string) string {
	langMap := map[string]string{
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".jsx":   "JavaScript",
		".tsx":   "TypeScript",
		".java":  "Java",
		".rs":    "Rust",
		".rb":    "Ruby",
		".php":   "PHP",
		".c":     "C",
		".cpp":   "C++",
		".cc":    "C++",
		".h":     "C/C++",
		".hpp":   "C++",
		".cs":    "C#",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",
		".ex":    "Elixir",
		".exs":   "Elixir",
		".erl":   "Erlang",
		".hs":    "Haskell",
		".lua":   "Lua",
		".r":     "R",
		".jl":    "Julia",
		".dart":  "Dart",
	}

	return langMap[ext]
}

// countLines counts the number of lines in a file
func (a *Analyzer) countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	return lineCount, scanner.Err()
}

// parseDependencies parses dependency files
func (a *Analyzer) parseDependencies(files []string, result *models.AnalysisResult) error {
	for _, file := range files {
		deps, lang, err := a.parserRegistry.ParseFile(file)
		if err != nil {
			// Not a dependency file or parsing error - skip
			continue
		}

		result.Dependencies = append(result.Dependencies, deps...)

		// Update language version if detected
		if lang != "" {
			a.updateLanguageVersion(file, lang, result)
		}
	}

	return nil
}

// updateLanguageVersion attempts to detect language version
func (a *Analyzer) updateLanguageVersion(filePath, lang string, result *models.AnalysisResult) {
	// Read first few lines to find version hints
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() && lineCount < 10 {
		line := scanner.Text()

		// Look for version patterns
		if lang == "Go" {
			if strings.HasPrefix(line, "go ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					for i, info := range result.Languages {
						if info.Name == "Go" {
							result.Languages[i].Version = parts[1]
							return
						}
					}
				}
			}
		} else if lang == "Python" {
			// Look for python_requires in pyproject.toml or similar
			re := regexp.MustCompile(`python.*["><=]+\s*["']?([0-9.]+)`)
			if matches := re.FindStringSubmatch(line); len(matches) >= 2 {
				for i, info := range result.Languages {
					if info.Name == "Python" {
						result.Languages[i].Version = matches[1]
						return
					}
				}
			}
		} else if lang == "JavaScript" || lang == "TypeScript" {
			// Look for Node.js version in package.json engines
			re := regexp.MustCompile(`"node":\s*"([^"]+)"`)
			if matches := re.FindStringSubmatch(line); len(matches) >= 2 {
				for i, info := range result.Languages {
					if info.Name == lang {
						result.Languages[i].Version = matches[1]
						return
					}
				}
			}
		}

		lineCount++
	}
}

// detectFramework detects web frameworks and their configurations
func (a *Analyzer) detectFramework(files []string, result *models.AnalysisResult) error {
	frameworks := map[string][]string{
		"Express":     {"express"},
		"FastAPI":     {"fastapi"},
		"Django":      {"django"},
		"Flask":       {"flask"},
		"Gin":         {"github.com/gin-gonic/gin"},
		"Fiber":       {"github.com/gofiber/fiber"},
		"Spring Boot": {"spring-boot"},
		"Rails":       {"rails"},
		"Laravel":     {"laravel/framework"},
		"Next.js":     {"next"},
		"Nuxt.js":     {"nuxt"},
		"Vue.js":      {"vue"},
		"React":       {"react"},
		"Angular":     {"@angular/core"},
	}

	// Check dependencies for framework indicators
	for _, dep := range result.Dependencies {
		for framework, indicators := range frameworks {
			for _, indicator := range indicators {
				if strings.Contains(dep.Name, indicator) {
					result.Framework = &models.FrameworkInfo{
						Name:    framework,
						Version: dep.Version,
						Type:    a.getFrameworkType(framework),
					}
					return nil
				}
			}
		}
	}

	return nil
}

// getFrameworkType determines the type of framework
func (a *Analyzer) getFrameworkType(framework string) string {
	webFrameworks := []string{"Express", "FastAPI", "Django", "Flask", "Gin", "Fiber", "Spring Boot", "Rails", "Laravel"}
	frontendFrameworks := []string{"React", "Vue.js", "Angular", "Next.js", "Nuxt.js"}

	for _, f := range webFrameworks {
		if f == framework {
			return "web"
		}
	}

	for _, f := range frontendFrameworks {
		if f == framework {
			return "frontend"
		}
	}

	return "unknown"
}

// findEntryPoints locates application entry points
func (a *Analyzer) findEntryPoints(files []string, result *models.AnalysisResult) error {
	entryPointPatterns := map[string][]string{
		"main.go":     {"Go"},
		"main.py":     {"Python"},
		"__main__.py": {"Python"},
		"app.py":      {"Python"},
		"server.py":   {"Python"},
		"index.js":    {"JavaScript"},
		"server.js":   {"JavaScript"},
		"app.js":      {"JavaScript"},
		"main.js":     {"JavaScript"},
		"index.ts":    {"TypeScript"},
		"server.ts":   {"TypeScript"},
		"Main.java":   {"Java"},
		"main.rs":     {"Rust"},
	}

	for _, file := range files {
		filename := filepath.Base(file)

		if langs, exists := entryPointPatterns[filename]; exists {
			entryPoint := models.EntryPoint{
				Path:     file,
				Type:     "main",
				Language: langs[0],
			}
			result.EntryPoints = append(result.EntryPoints, entryPoint)
		}
	}

	return nil
}

// extractConfiguration extracts ports, environment variables, and other config
func (a *Analyzer) extractConfiguration(files []string, result *models.AnalysisResult) error {
	portRegex := regexp.MustCompile(`(?i)(port|PORT).*?[:=]\s*(\d+)`)

	for _, file := range files {
		// Only check config files and main entry points
		filename := filepath.Base(file)
		if !a.isConfigFile(filename) && !a.isEntryPoint(filename) {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Extract ports
		matches := portRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				port := match[2]
				if p := a.parseInt(port); p > 0 && p < 65536 {
					result.Ports = append(result.Ports, p)
				}
			}
		}
	}

	// Remove duplicate ports
	result.Ports = a.uniqueInts(result.Ports)

	return nil
}

// isConfigFile checks if a file is a configuration file
func (a *Analyzer) isConfigFile(filename string) bool {
	configFiles := []string{
		"config.yaml", "config.yml", "config.json", "config.toml",
		".env", ".env.example", "settings.py", "application.properties",
		"application.yml", "application.yaml",
	}

	for _, cf := range configFiles {
		if filename == cf {
			return true
		}
	}

	return false
}

// isEntryPoint checks if a file is an entry point
func (a *Analyzer) isEntryPoint(filename string) bool {
	entryPoints := []string{
		"main.go", "main.py", "server.py", "app.py",
		"index.js", "server.js", "app.js", "main.js",
		"Main.java", "main.rs",
	}

	for _, ep := range entryPoints {
		if filename == ep {
			return true
		}
	}

	return false
}

// calculateFileStats calculates various file statistics
func (a *Analyzer) calculateFileStats(files []string, result *models.AnalysisResult) error {
	result.FileStats.TotalFiles = len(files)
	maxSize := int64(0)
	var largestFile string

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.Size() > maxSize {
			maxSize = info.Size()
			largestFile = file
		}

		// Count language distribution
		ext := filepath.Ext(file)
		lang := a.getLanguageFromExtension(ext)
		if lang != "" {
			result.FileStats.FilesByLanguage[lang]++
		}
	}

	result.FileStats.LargestFile = largestFile
	result.FileStats.LargestFileSize = maxSize

	return nil
}

// determinePrimaryLanguage determines the primary language of the project
func (a *Analyzer) determinePrimaryLanguage(result *models.AnalysisResult) {
	if len(result.Languages) == 0 {
		return
	}

	// Find language with highest percentage
	maxPercentage := 0.0
	primaryLang := ""

	for _, lang := range result.Languages {
		if lang.Percentage > maxPercentage {
			maxPercentage = lang.Percentage
			primaryLang = lang.Name
		}
	}

	result.PrimaryLanguage = primaryLang
}

// detectBuildTools detects package managers and build tools
func (a *Analyzer) detectBuildTools(files []string, result *models.AnalysisResult) {
	for _, file := range files {
		filename := filepath.Base(file)

		switch filename {
		case "package.json":
			result.PackageManager = "npm"
			result.BuildTool = "npm"
		case "yarn.lock":
			result.PackageManager = "yarn"
		case "pnpm-lock.yaml":
			result.PackageManager = "pnpm"
		case "go.mod":
			result.PackageManager = "go modules"
			result.BuildTool = "go"
		case "requirements.txt", "pyproject.toml":
			result.PackageManager = "pip"
			result.BuildTool = "python"
		case "Cargo.toml":
			result.PackageManager = "cargo"
			result.BuildTool = "cargo"
		case "Gemfile":
			result.PackageManager = "bundler"
			result.BuildTool = "bundler"
		case "pom.xml":
			result.PackageManager = "maven"
			result.BuildTool = "maven"
		case "build.gradle", "build.gradle.kts":
			result.PackageManager = "gradle"
			result.BuildTool = "gradle"
		case "Makefile":
			result.BuildTool = "make"
		}
	}
}

// Helper functions

func (a *Analyzer) parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

func (a *Analyzer) uniqueInts(slice []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, val := range slice {
		if !seen[val] {
			seen[val] = true
			result = append(result, val)
		}
	}

	return result
}
