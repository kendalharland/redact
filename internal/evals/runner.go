// Package evals implements the evaluation system for the redact tool.
package evals

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/kendalharland/redact/internal/config"
	"github.com/kendalharland/redact/internal/matcher"
	"github.com/kendalharland/redact/internal/pipeline"
)

// Options holds the evaluation options.
type Options struct {
	Paths      []string
	JSONOutput string
	Model      string
}

// ParseFlags parses the evaluation command flags.
func ParseFlags(args []string) (*Options, error) {
	fs := flag.NewFlagSet("evals", flag.ContinueOnError)

	opts := &Options{}
	fs.StringVar(&opts.JSONOutput, "json-output", "", "Write metrics to this JSON file instead of stdout")
	fs.StringVar(&opts.Model, "model", "", "Claude model to use (default: claude-sonnet-4-20250514)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Remaining arguments are paths
	opts.Paths = fs.Args()

	return opts, nil
}

// Runner runs evaluations.
type Runner struct {
	cfg      *config.Config
	client   *config.Client
	registry *matcher.Registry
}

// NewRunner creates a new evaluation runner.
func NewRunner(model string) *Runner {
	cfg := config.LoadConfig()
	if model != "" {
		cfg.Model = model
	}
	client := config.NewClient(cfg)

	return &Runner{
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

// Run executes the evaluation.
func (r *Runner) Run(opts *Options, stdout, stderr io.Writer) error {
	// Find evaluation files
	paths, err := r.findEvalFiles(opts.Paths)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		return fmt.Errorf("no evaluation files found")
	}

	// Generate run ID
	runID := uuid.New().String()

	// Run evaluations and collect metrics
	var allResults []TestCaseResult
	for _, inputPath := range paths {
		result, err := r.runEval(inputPath, runID, stderr)
		if err != nil {
			fmt.Fprintf(stderr, "Warning: failed to run eval for %s: %v\n", inputPath, err)
			continue
		}
		allResults = append(allResults, result)
	}

	// Compute aggregate metrics
	aggregateMetrics := ComputeAggregateMetrics(allResults)

	// Generate report
	report := &Report{
		RunID:            runID,
		TestCaseResults:  allResults,
		AggregateMetrics: aggregateMetrics,
	}

	// Output report
	if opts.JSONOutput != "" {
		return r.writeJSONReport(report, opts.JSONOutput)
	}

	PrintReport(report, stdout)
	return nil
}

// findEvalFiles finds all evaluation input files.
func (r *Runner) findEvalFiles(paths []string) ([]string, error) {
	if len(paths) == 0 {
		// Default: look in evals/ directory
		return filepath.Glob("evals/*.txt")
	}

	var result []string
	for _, p := range paths {
		// Handle glob patterns
		matches, err := filepath.Glob(p)
		if err != nil {
			return nil, err
		}
		result = append(result, matches...)
	}

	return result, nil
}

// runEval runs evaluation for a single test case.
func (r *Runner) runEval(inputPath string, runID string, stderr io.Writer) (TestCaseResult, error) {
	testCaseName := strings.TrimSuffix(filepath.Base(inputPath), ".txt")

	// Read input text
	text, err := os.ReadFile(inputPath)
	if err != nil {
		return TestCaseResult{}, err
	}

	// Read ground truth labels
	labelsPath := strings.TrimSuffix(inputPath, ".txt") + ".labels.json"
	groundTruth, err := r.loadGroundTruth(labelsPath)
	if err != nil {
		return TestCaseResult{}, fmt.Errorf("failed to load ground truth: %w", err)
	}

	// Run the pipeline
	p := pipeline.NewPipeline(string(text))

	matchers := r.registry.GetAll()
	for _, m := range matchers {
		matches, err := m.Match(string(text))
		if err != nil {
			fmt.Fprintf(stderr, "Warning: %s matcher failed for %s: %v\n", m.Type(), testCaseName, err)
			continue
		}
		p.AddMatches(m.Type(), matches)
	}

	// Add obfuscated matcher if API key is available
	if r.cfg.HasAPIKey {
		obfuscatedMatcher := matcher.NewObfuscatedMatcher(r.client, pipeline.AllPatternTypes())
		matches, err := obfuscatedMatcher.Match(string(text))
		if err != nil {
			fmt.Fprintf(stderr, "Warning: Obfuscated matcher failed for %s: %v\n", testCaseName, err)
		} else {
			p.AddMatches(pipeline.Obfuscated, matches)
		}
	}

	p.Process()

	// Get generated labels
	generatedLabels := p.GetLabels()

	// Compute metrics
	inputText := string(text)
	metrics := ComputeMetrics(generatedLabels, groundTruth, inputText, testCaseName)

	// Save results
	if err := r.saveResults(testCaseName, runID, p); err != nil {
		fmt.Fprintf(stderr, "Warning: failed to save results for %s: %v\n", testCaseName, err)
	}

	return TestCaseResult{
		TestCaseName:    testCaseName,
		InputText:       inputText,
		Metrics:         metrics,
		GeneratedLabels: generatedLabels,
		GroundTruth:     groundTruth,
	}, nil
}

// loadGroundTruth loads ground truth labels from a JSON file.
func (r *Runner) loadGroundTruth(path string) (pipeline.Labels, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return pipeline.LabelsFromJSON(data)
}

// saveResults saves the evaluation results to disk.
func (r *Runner) saveResults(testCaseName, runID string, p *pipeline.Pipeline) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	resultsDir := filepath.Join(homeDir, ".config", "redact", "evals", testCaseName, runID)
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return err
	}

	// Save classification map
	cmapJSON, err := p.GetClassificationMap().ToJSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(resultsDir, "classmap.json"), cmapJSON, 0644); err != nil {
		return err
	}

	// Save labels
	labelsJSON, err := p.GetLabels().ToJSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(resultsDir, "labels.json"), labelsJSON, 0644); err != nil {
		return err
	}

	// Save redacted output
	redacted := p.GetRedactedText(pipeline.RedactOptions{})
	if err := os.WriteFile(filepath.Join(resultsDir, "output.txt"), []byte(redacted), 0644); err != nil {
		return err
	}

	return nil
}

// writeJSONReport writes the report as JSON to a file.
func (r *Runner) writeJSONReport(report *Report, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
