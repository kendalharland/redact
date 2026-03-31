// Package config handles configuration and API client setup.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	// AnthropicAPIKeyEnv is the environment variable for the Anthropic API key.
	AnthropicAPIKeyEnv = "ANTHROPIC_API_KEY"

	// AnthropicAPIURL is the base URL for the Anthropic API.
	AnthropicAPIURL = "https://api.anthropic.com/v1/messages"

	// DefaultModel is the default Claude model to use.
	DefaultModel = "claude-sonnet-4-20250514"
)

// Config holds the application configuration.
type Config struct {
	APIKey    string
	Model     string
	HasAPIKey bool
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	apiKey := os.Getenv(AnthropicAPIKeyEnv)
	return &Config{
		APIKey:    apiKey,
		Model:     DefaultModel,
		HasAPIKey: apiKey != "",
	}
}

// Client provides access to the Anthropic API.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Anthropic API client.
func NewClient(config *Config) *Client {
	return &Client{
		config:     config,
		httpClient: &http.Client{},
	}
}

// Message represents a message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// APIRequest represents a request to the Anthropic API.
type APIRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// ContentBlock represents a content block in the API response.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// APIResponse represents a response from the Anthropic API.
type APIResponse struct {
	Content []ContentBlock `json:"content"`
	Error   *APIError      `json:"error,omitempty"`
}

// APIError represents an error from the Anthropic API.
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Complete sends a prompt to the Anthropic API and returns the response text.
func (c *Client) Complete(prompt string) (string, error) {
	if !c.config.HasAPIKey {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	request := APIRequest{
		Model:     c.config.Model,
		MaxTokens: 4096,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", AnthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s - %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	// Concatenate all text content blocks
	var result strings.Builder
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}

	return result.String(), nil
}
