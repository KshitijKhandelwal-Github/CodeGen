package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/company/codegen-pro/internal/models"
)

// Client handles AI API interactions
type Client struct {
	Provider string
	Model    string
	APIKey   string
	client   *http.Client
}

// NewClient creates a new AI client
func NewClient(provider, model, apiKey string) (*Client, error) {
	if apiKey == "" {
		switch provider {
		case "anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "gemini":
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
	}

	if apiKey == "" {
		envVar := map[string]string{
			"anthropic": "ANTHROPIC",
			"openai":    "OPENAI",
			"gemini":    "GEMINI",
		}[provider]
		return nil, fmt.Errorf("no API key provided. Set api_key in .codegen.yaml or use environment variable (%s_API_KEY)", envVar)
	}

	if provider != "anthropic" && provider != "openai" && provider != "gemini" {
		return nil, fmt.Errorf("unsupported AI provider: %s (supported: anthropic, openai, gemini)", provider)
	}

	return &Client{
		Provider: provider,
		Model:    model,
		APIKey:   apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// EnhanceREADME uses AI to analyze the codebase and generate comprehensive description with features
func (c *Client) EnhanceREADME(analysis *models.AnalysisResult, config *models.ProjectConfig) (string, error) {
	// Read key files from the codebase
	codebaseContext, err := c.readCodebaseFiles(analysis)
	if err != nil {
		return "", fmt.Errorf("error reading codebase: %w", err)
	}

	prompt := c.buildEnhancedPrompt(analysis, config, codebaseContext)

	switch c.Provider {
	case "anthropic":
		return c.callAnthropic(prompt)
	case "openai":
		return c.callOpenAI(prompt)
	case "gemini":
		return c.callGemini(prompt)
	default:
		return "", fmt.Errorf("unsupported provider: %s", c.Provider)
	}
}

// readCodebaseFiles intelligently selects and reads key files from the codebase
func (c *Client) readCodebaseFiles(analysis *models.AnalysisResult) (string, error) {
	var codeSnippets []string
	projectPath := analysis.ProjectPath

	// Priority files to read (in order of importance)
	priorityFiles := []string{
		// Documentation
		"README.md", "README.txt", "OVERVIEW.md",
		// Entry points
		"main.go", "main.py", "index.js", "app.py", "server.js", "index.ts",
		"cmd/main.go", "src/main.go", "src/index.js", "src/App.js",
		// Configuration
		"package.json", "go.mod", "Cargo.toml", "pyproject.toml", "requirements.txt",
		"Makefile", "Dockerfile", "docker-compose.yml",
	}

	// Add detected entry points
	for _, ep := range analysis.EntryPoints {
		priorityFiles = append(priorityFiles, ep.Path)
	}

	filesRead := 0
	maxFiles := 10
	maxFileSize := 5000

	for _, filename := range priorityFiles {
		if filesRead >= maxFiles {
			break
		}

		fullPath := filepath.Join(projectPath, filename)
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if len(contentStr) > maxFileSize {
			contentStr = contentStr[:maxFileSize] + "\n... (truncated)"
		}

		codeSnippets = append(codeSnippets, fmt.Sprintf("=== %s ===\n%s\n", filename, contentStr))
		filesRead++
	}

	if len(codeSnippets) == 0 {
		return "No code files available for analysis.", nil
	}

	return strings.Join(codeSnippets, "\n"), nil
}

// buildEnhancedPrompt creates a comprehensive prompt for codebase analysis
func (c *Client) buildEnhancedPrompt(analysis *models.AnalysisResult, config *models.ProjectConfig, codebaseContext string) string {
	prompt := fmt.Sprintf(`You are analyzing a software project to generate a comprehensive README overview. Here's what you know:

PROJECT INFORMATION:
- Name: %s
- Primary Language: %s`, config.ProjectName, analysis.PrimaryLanguage)

	if analysis.Framework != nil {
		prompt += fmt.Sprintf("\n- Framework: %s %s (%s)", analysis.Framework.Name, analysis.Framework.Version, analysis.Framework.Type)
	}

	if len(analysis.Languages) > 0 {
		prompt += "\n\nLANGUAGES USED:\n"
		for _, lang := range analysis.Languages {
			prompt += fmt.Sprintf("- %s: %.1f%% (%d files)\n", lang.Name, lang.Percentage, lang.FileCount)
		}
	}

	if len(analysis.Dependencies) > 0 {
		prompt += fmt.Sprintf("\n\nKEY DEPENDENCIES (%d total):\n", len(analysis.Dependencies))
		count := 0
		for _, dep := range analysis.Dependencies {
			if count >= 15 {
				break
			}
			if dep.Type == "runtime" {
				prompt += fmt.Sprintf("- %s (%s)\n", dep.Name, dep.Version)
				count++
			}
		}
	}

	if len(analysis.Ports) > 0 {
		prompt += fmt.Sprintf("\n\nEXPOSED PORTS: %v\n", analysis.Ports)
	}

	prompt += fmt.Sprintf("\n\nCODEBASE FILES:\n%s\n", codebaseContext)

	prompt += `

TASK:
Based on the project information and code files above, generate a comprehensive project overview (3-5 paragraphs) that includes:

1. WHAT THE PROJECT DOES: Clearly explain the primary purpose and functionality based on the actual code
2. KEY FEATURES: List 4-8 specific features you can identify from the codebase (be specific, not generic)
3. TECHNICAL HIGHLIGHTS: Notable technical decisions, architecture patterns, or implementation details
4. USE CASES: Who would use this and what problems it solves

IMPORTANT:
- Be SPECIFIC based on what you see in the code - don't be generic
- Identify actual features implemented in the codebase
- Mention specific technologies, patterns, or approaches used
- Write in professional, engaging language suitable for a GitHub README
- Return ONLY plain text (no markdown formatting)
- Structure: First paragraph = overview, Second paragraph = key features, Third paragraph = technical highlights and use cases`

	return prompt
}

// callAnthropic calls the Anthropic Claude API
func (c *Client) callAnthropic(prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	reqBody := map[string]interface{}{
		"model":      c.Model,
		"max_tokens": 2000,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return result.Content[0].Text, nil
}

// callOpenAI calls the OpenAI API
func (c *Client) callOpenAI(prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  2000,
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

// callGemini calls the Google Gemini API
func (c *Client) callGemini(prompt string) (string, error) {
	// Gemini API endpoint
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 2000,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling Gemini API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}