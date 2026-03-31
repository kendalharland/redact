package evals

import (
	"github.com/kendalharland/redact/internal/pipeline"
)

// MisclassifiedItem holds information about a false positive or false negative.
type MisclassifiedItem struct {
	TestCase string `json:"test_case"`
	Offset   int    `json:"offset"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

// Metrics holds precision and recall metrics.
type Metrics struct {
	Precision       float64                `json:"precision"`
	Recall          float64                `json:"recall"`
	TruePositives   int                    `json:"true_positives"`
	FalsePositives  int                    `json:"false_positives"`
	FalseNegatives  int                    `json:"false_negatives"`
	ByType          map[string]TypeMetrics `json:"by_type"`
	FalsePositiveItems []MisclassifiedItem `json:"false_positives_detail,omitempty"`
	FalseNegativeItems []MisclassifiedItem `json:"false_negatives_detail,omitempty"`
}

// TypeMetrics holds metrics for a specific pattern type.
type TypeMetrics struct {
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	FalseNegatives int     `json:"false_negatives"`
}

// TestCaseResult holds the result of a single test case.
type TestCaseResult struct {
	TestCaseName    string           `json:"test_case_name"`
	InputText       string           `json:"-"` // Not included in JSON output
	Metrics         Metrics          `json:"metrics"`
	GeneratedLabels pipeline.Labels  `json:"generated_labels,omitempty"`
	GroundTruth     pipeline.Labels  `json:"ground_truth,omitempty"`
}

// ComputeMetrics computes precision and recall metrics by comparing generated labels
// to ground truth labels.
func ComputeMetrics(generated, groundTruth pipeline.Labels, inputText, testCaseName string) Metrics {
	metrics := Metrics{
		ByType: make(map[string]TypeMetrics),
	}

	// Build lookup maps
	genMap := buildLabelMap(generated)
	gtMap := buildLabelMap(groundTruth)

	// Compute true positives and false positives
	for key, genLabel := range genMap {
		baseType := pipeline.ExtractBaseType(genLabel.Type)
		tm := metrics.ByType[baseType]

		if _, found := gtMap[key]; found {
			metrics.TruePositives++
			tm.TruePositives++
		} else {
			metrics.FalsePositives++
			tm.FalsePositives++
			metrics.FalsePositiveItems = append(metrics.FalsePositiveItems, MisclassifiedItem{
				TestCase: testCaseName,
				Offset:   genLabel.Offset,
				Text:     extractText(inputText, genLabel.Offset, genLabel.Length),
				Type:     genLabel.Type,
			})
		}
		metrics.ByType[baseType] = tm
	}

	// Compute false negatives
	for key, gtLabel := range gtMap {
		baseType := pipeline.ExtractBaseType(gtLabel.Type)
		tm := metrics.ByType[baseType]

		if _, found := genMap[key]; !found {
			metrics.FalseNegatives++
			tm.FalseNegatives++
			metrics.FalseNegativeItems = append(metrics.FalseNegativeItems, MisclassifiedItem{
				TestCase: testCaseName,
				Offset:   gtLabel.Offset,
				Text:     extractText(inputText, gtLabel.Offset, gtLabel.Length),
				Type:     gtLabel.Type,
			})
		}
		metrics.ByType[baseType] = tm
	}

	// Compute precision and recall
	metrics.Precision = computePrecision(metrics.TruePositives, metrics.FalsePositives)
	metrics.Recall = computeRecall(metrics.TruePositives, metrics.FalseNegatives)

	// Compute per-type precision and recall
	for typeName, tm := range metrics.ByType {
		tm.Precision = computePrecision(tm.TruePositives, tm.FalsePositives)
		tm.Recall = computeRecall(tm.TruePositives, tm.FalseNegatives)
		metrics.ByType[typeName] = tm
	}

	return metrics
}

// extractText safely extracts a substring from text given offset and length.
func extractText(text string, offset, length int) string {
	if offset < 0 || offset >= len(text) {
		return "<out of bounds>"
	}
	end := offset + length
	if end > len(text) {
		end = len(text)
	}
	return text[offset:end]
}

// labelKey generates a unique key for a label based on offset and length.
// We use offset+length to identify labels since the type might differ
// (e.g., PERSON_1 vs PERSON_2 for the same position).
type labelKey struct {
	offset int
	length int
}

// buildLabelMap creates a map from label positions to labels for quick lookup.
func buildLabelMap(labels pipeline.Labels) map[labelKey]pipeline.Label {
	m := make(map[labelKey]pipeline.Label)
	for _, label := range labels {
		key := labelKey{offset: label.Offset, length: label.Length}
		m[key] = label
	}
	return m
}

// computePrecision computes precision from TP and FP counts.
func computePrecision(tp, fp int) float64 {
	if tp+fp == 0 {
		return 1.0 // No predictions made
	}
	return float64(tp) / float64(tp+fp)
}

// computeRecall computes recall from TP and FN counts.
func computeRecall(tp, fn int) float64 {
	if tp+fn == 0 {
		return 1.0 // No ground truth labels
	}
	return float64(tp) / float64(tp+fn)
}

// ComputeAggregateMetrics computes aggregate metrics from multiple test case results.
func ComputeAggregateMetrics(results []TestCaseResult) Metrics {
	aggregate := Metrics{
		ByType: make(map[string]TypeMetrics),
	}

	for _, result := range results {
		aggregate.TruePositives += result.Metrics.TruePositives
		aggregate.FalsePositives += result.Metrics.FalsePositives
		aggregate.FalseNegatives += result.Metrics.FalseNegatives

		for typeName, tm := range result.Metrics.ByType {
			atm := aggregate.ByType[typeName]
			atm.TruePositives += tm.TruePositives
			atm.FalsePositives += tm.FalsePositives
			atm.FalseNegatives += tm.FalseNegatives
			aggregate.ByType[typeName] = atm
		}

		aggregate.FalsePositiveItems = append(aggregate.FalsePositiveItems, result.Metrics.FalsePositiveItems...)
		aggregate.FalseNegativeItems = append(aggregate.FalseNegativeItems, result.Metrics.FalseNegativeItems...)
	}

	// Compute aggregate precision and recall
	aggregate.Precision = computePrecision(aggregate.TruePositives, aggregate.FalsePositives)
	aggregate.Recall = computeRecall(aggregate.TruePositives, aggregate.FalseNegatives)

	// Compute per-type aggregate precision and recall
	for typeName, tm := range aggregate.ByType {
		tm.Precision = computePrecision(tm.TruePositives, tm.FalsePositives)
		tm.Recall = computeRecall(tm.TruePositives, tm.FalseNegatives)
		aggregate.ByType[typeName] = tm
	}

	return aggregate
}
