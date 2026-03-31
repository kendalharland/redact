package evals

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Report holds the complete evaluation report.
type Report struct {
	RunID            string            `json:"run_id"`
	TestCaseResults  []TestCaseResult  `json:"test_case_results"`
	AggregateMetrics Metrics           `json:"aggregate_metrics"`
}

// PrintReport prints the evaluation report to the writer.
func PrintReport(report *Report, w io.Writer) {
	fmt.Fprintf(w, "Evaluation Report\n")
	fmt.Fprintf(w, "=================\n")
	fmt.Fprintf(w, "Run ID: %s\n\n", report.RunID)

	// Print aggregate metrics
	fmt.Fprintf(w, "Overall Metrics\n")
	fmt.Fprintf(w, "---------------\n")
	printMetricsTable(w, report.AggregateMetrics, "")

	// Print per-test-case metrics
	if len(report.TestCaseResults) > 1 {
		fmt.Fprintf(w, "\nPer-Test-Case Metrics\n")
		fmt.Fprintf(w, "---------------------\n")
		for _, result := range report.TestCaseResults {
			fmt.Fprintf(w, "\n%s:\n", result.TestCaseName)
			printMetricsTable(w, result.Metrics, "  ")
		}
	}

	// Print false positives and false negatives
	if len(report.AggregateMetrics.FalsePositiveItems) > 0 {
		fmt.Fprintf(w, "\nFalse Positives\n")
		fmt.Fprintf(w, "---------------\n")
		for _, item := range report.AggregateMetrics.FalsePositiveItems {
			fmt.Fprintf(w, "  [%s] offset %d: %q (%s)\n", item.TestCase, item.Offset, item.Text, item.Type)
		}
	}

	if len(report.AggregateMetrics.FalseNegativeItems) > 0 {
		fmt.Fprintf(w, "\nFalse Negatives (Missed)\n")
		fmt.Fprintf(w, "------------------------\n")
		for _, item := range report.AggregateMetrics.FalseNegativeItems {
			fmt.Fprintf(w, "  [%s] offset %d: %q (%s)\n", item.TestCase, item.Offset, item.Text, item.Type)
		}
	}
}

// printMetricsTable prints a metrics table.
func printMetricsTable(w io.Writer, metrics Metrics, indent string) {
	// Header
	fmt.Fprintf(w, "%s%-20s %10s %10s %6s %6s %6s\n",
		indent, "Type", "Precision", "Recall", "TP", "FP", "FN")
	fmt.Fprintf(w, "%s%s\n", indent, strings.Repeat("-", 70))

	// Overall
	fmt.Fprintf(w, "%s%-20s %10.2f%% %10.2f%% %6d %6d %6d\n",
		indent, "OVERALL",
		metrics.Precision*100, metrics.Recall*100,
		metrics.TruePositives, metrics.FalsePositives, metrics.FalseNegatives)

	// Per-type (sorted)
	var typeNames []string
	for typeName := range metrics.ByType {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)

	for _, typeName := range typeNames {
		tm := metrics.ByType[typeName]
		fmt.Fprintf(w, "%s%-20s %10.2f%% %10.2f%% %6d %6d %6d\n",
			indent, typeName,
			tm.Precision*100, tm.Recall*100,
			tm.TruePositives, tm.FalsePositives, tm.FalseNegatives)
	}
}
