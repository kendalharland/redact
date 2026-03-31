// Package cli implements the command-line interface for redact.
package cli

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kendalharland/redact/internal/config"
	"github.com/kendalharland/redact/internal/matcher"
	"github.com/kendalharland/redact/internal/pipeline"
)

// Options holds the CLI options.
type Options struct {
	Bleep        bool
	Exclude      string
	OutputCmap   string
	OutputLabels string
	Model        string
	InputFile    string
}

// App represents the CLI application.
type App struct {
	cfg      *config.Config
	client   *config.Client
	registry *matcher.Registry
}

// NewApp creates a new CLI application.
func NewApp(model string) *App {
	cfg := config.LoadConfig()
	if model != "" {
		cfg.Model = model
	}
	client := config.NewClient(cfg)

	return &App{
		cfg:      cfg,
		client:   client,
		registry: buildRegistry(client, cfg.HasAPIKey),
	}
}

// buildRegistry creates a matcher registry with all available matchers.
func buildRegistry(client *config.Client, hasAPIKey bool) *matcher.Registry {
	reg := matcher.NewRegistry()

	// Register pattern matchers (always available)
	reg.Register(matcher.NewCreditCardMatcher())
	reg.Register(matcher.NewEmailMatcher())
	reg.Register(matcher.NewIPAddressMatcher())
	reg.Register(matcher.NewMACAddressMatcher())
	reg.Register(matcher.NewPhoneMatcher())

	// Register LLM matchers (only if API key is available)
	if hasAPIKey {
		reg.Register(matcher.NewPersonMatcher(client))
		// Obfuscated matcher will be created with enabled types later
	}

	return reg
}

// ParseFlags parses command-line flags and returns options.
func ParseFlags(args []string) (*Options, error) {
	fs := flag.NewFlagSet("redact", flag.ContinueOnError)

	opts := &Options{}
	fs.BoolVar(&opts.Bleep, "bleep", false, "Replace matches with *** instead of placeholders")
	fs.StringVar(&opts.Exclude, "exclude", "", "Comma-separated list of types to exclude (e.g., phone,email)")
	fs.StringVar(&opts.OutputCmap, "output-cmap", "", "Write classification map to this file")
	fs.StringVar(&opts.OutputLabels, "output-labels", "", "Write labels to this file")
	fs.StringVar(&opts.Model, "model", "", "Claude model to use (default: claude-sonnet-4-20250514)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Get positional argument (input file)
	if fs.NArg() > 0 {
		opts.InputFile = fs.Arg(0)
	}

	return opts, nil
}

// parseExcludeTypes parses the exclude flag value into pattern types.
func parseExcludeTypes(exclude string) ([]pipeline.PatternType, error) {
	if exclude == "" {
		return nil, nil
	}

	var excluded []pipeline.PatternType
	parts := strings.Split(exclude, ",")
	for _, part := range parts {
		pt, err := pipeline.PatternTypeFromString(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		excluded = append(excluded, pt)
	}
	return excluded, nil
}

// Run executes the main redaction logic.
func (a *App) Run(opts *Options, stdin io.Reader, stdout, stderr io.Writer) error {
	// Parse exclude types
	excludeTypes, err := parseExcludeTypes(opts.Exclude)
	if err != nil {
		return fmt.Errorf("invalid exclude types: %w", err)
	}

	// Read input
	text, err := readInput(opts.InputFile, stdin)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Warn if LLM matchers are disabled
	if !a.cfg.HasAPIKey {
		fmt.Fprintln(stderr, "Warning: ANTHROPIC_API_KEY not set. Person and Obfuscated matchers are disabled.")
	}

	// Get enabled matchers
	enabledMatchers := a.registry.GetEnabled(excludeTypes)

	// Create pipeline
	p := pipeline.NewPipeline(text)

	// Run all matchers
	for _, m := range enabledMatchers {
		matches, err := m.Match(text)
		if err != nil {
			fmt.Fprintf(stderr, "Warning: %s matcher failed: %v\n", m.Type(), err)
			continue
		}
		p.AddMatches(m.Type(), matches)
	}

	// Add obfuscated matcher if API key is available and not excluded
	if a.cfg.HasAPIKey && !isExcluded(excludeTypes, pipeline.Obfuscated) {
		// Get enabled patterned types for obfuscated detection
		var enabledTypes []pipeline.PatternType
		for _, pt := range pipeline.AllPatternTypes() {
			if !isExcluded(excludeTypes, pt) {
				enabledTypes = append(enabledTypes, pt)
			}
		}

		obfuscatedMatcher := matcher.NewObfuscatedMatcher(a.client, enabledTypes)
		matches, err := obfuscatedMatcher.Match(text)
		if err != nil {
			fmt.Fprintf(stderr, "Warning: Obfuscated matcher failed: %v\n", err)
		} else {
			p.AddMatches(pipeline.Obfuscated, matches)
		}
	}

	// Process pipeline
	p.Process()

	// Write classification map if requested
	if opts.OutputCmap != "" {
		cmapJSON, err := p.GetClassificationMap().ToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize classification map: %w", err)
		}
		if err := os.WriteFile(opts.OutputCmap, cmapJSON, 0644); err != nil {
			return fmt.Errorf("failed to write classification map: %w", err)
		}
	}

	// Write labels if requested
	if opts.OutputLabels != "" {
		labelsJSON, err := p.GetLabels().ToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize labels: %w", err)
		}
		if err := os.WriteFile(opts.OutputLabels, labelsJSON, 0644); err != nil {
			return fmt.Errorf("failed to write labels: %w", err)
		}
	}

	// Output redacted text
	redacted := p.GetRedactedText(pipeline.RedactOptions{Bleep: opts.Bleep})
	fmt.Fprint(stdout, redacted)

	return nil
}

// readInput reads the input text from a file or stdin.
func readInput(filePath string, stdin io.Reader) (string, error) {
	var reader io.Reader

	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer f.Close()
		reader = f
	} else {
		reader = stdin
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // Support large inputs

	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Remove trailing newline if input was a single line
	result := sb.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// isExcluded checks if a pattern type is in the excluded list.
func isExcluded(excludeTypes []pipeline.PatternType, pt pipeline.PatternType) bool {
	for _, excluded := range excludeTypes {
		if excluded == pt {
			return true
		}
	}
	return false
}
