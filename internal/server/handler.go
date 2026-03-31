// Package server implements the web server for the redact tool.
package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kendalharland/redact/internal/config"
	"github.com/kendalharland/redact/internal/matcher"
	"github.com/kendalharland/redact/internal/pipeline"
)

// RedactRequest is the JSON request body for the redact endpoint.
type RedactRequest struct {
	Text     string   `json:"text"`
	Types    []string `json:"types"`
	BleepMode bool    `json:"bleep_mode"`
}

// RedactResponse is the JSON response body for the redact endpoint.
type RedactResponse struct {
	RedactedText string          `json:"redacted_text"`
	Labels       pipeline.Labels `json:"labels"`
	Error        string          `json:"error,omitempty"`
}

// Handler handles HTTP requests for the redact server.
type Handler struct {
	cfg      *config.Config
	client   *config.Client
	registry *matcher.Registry
}

// NewHandler creates a new Handler.
func NewHandler(model string) *Handler {
	cfg := config.LoadConfig()
	if model != "" {
		cfg.Model = model
	}
	client := config.NewClient(cfg)

	return &Handler{
		cfg:      cfg,
		client:   client,
		registry: buildRegistry(client, cfg.HasAPIKey),
	}
}

// buildRegistry creates a matcher registry with all available matchers.
func buildRegistry(client *config.Client, hasAPIKey bool) *matcher.Registry {
	reg := matcher.NewRegistry()

	// Register pattern matchers
	reg.Register(matcher.NewCreditCardMatcher())
	reg.Register(matcher.NewEmailMatcher())
	reg.Register(matcher.NewIPAddressMatcher())
	reg.Register(matcher.NewMACAddressMatcher())
	reg.Register(matcher.NewPhoneMatcher())

	// Register LLM matchers (only if API key is available)
	if hasAPIKey {
		reg.Register(matcher.NewPersonMatcher(client))
	}

	return reg
}

// HandleRedact handles POST /api/redact requests.
func (h *Handler) HandleRedact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RedactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, RedactResponse{Error: "Invalid JSON: " + err.Error()})
		return
	}

	if req.Text == "" {
		writeJSON(w, http.StatusBadRequest, RedactResponse{Error: "Text is required"})
		return
	}

	// Build set of enabled types
	enabledTypes := make(map[string]bool)
	for _, t := range req.Types {
		enabledTypes[strings.ToUpper(t)] = true
	}

	// Run the pipeline
	p := pipeline.NewPipeline(req.Text)

	matchers := h.registry.GetAll()
	for _, m := range matchers {
		// Skip if type not enabled
		typeName := string(m.Type())
		if len(enabledTypes) > 0 && !enabledTypes[typeName] {
			continue
		}

		matches, err := m.Match(req.Text)
		if err != nil {
			// Log but continue
			continue
		}
		p.AddMatches(m.Type(), matches)
	}

	// Add obfuscated matcher if API key is available and OBFUSCATED is enabled
	if h.cfg.HasAPIKey && (len(enabledTypes) == 0 || enabledTypes["OBFUSCATED"]) {
		obfuscatedMatcher := matcher.NewObfuscatedMatcher(h.client, pipeline.AllPatternTypes())
		matches, err := obfuscatedMatcher.Match(req.Text)
		if err == nil {
			p.AddMatches(pipeline.Obfuscated, matches)
		}
	}

	p.Process()

	// Get results
	labels := p.GetLabels()
	redacted := p.GetRedactedText(pipeline.RedactOptions{Bleep: req.BleepMode})

	writeJSON(w, http.StatusOK, RedactResponse{
		RedactedText: redacted,
		Labels:       labels,
	})
}

// HandleTypes returns the available pattern types.
func (h *Handler) HandleTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	types := []map[string]interface{}{
		{"name": "CREDIT_CARD", "label": "Credit Cards", "requiresLLM": false},
		{"name": "EMAIL_ADDRESS", "label": "Email Addresses", "requiresLLM": false},
		{"name": "IP_ADDRESS", "label": "IP Addresses", "requiresLLM": false},
		{"name": "MAC_ADDRESS", "label": "MAC Addresses", "requiresLLM": false},
		{"name": "PHONE_NUMBER", "label": "Phone Numbers", "requiresLLM": false},
		{"name": "PERSON", "label": "Person Names", "requiresLLM": true, "enabled": h.cfg.HasAPIKey},
		{"name": "OBFUSCATED", "label": "Obfuscated Data", "requiresLLM": true, "enabled": h.cfg.HasAPIKey},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"types":     types,
		"hasAPIKey": h.cfg.HasAPIKey,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
