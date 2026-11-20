package main

import (
	"slices"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/evidence"
)

type uiResult struct {
	ChatRef    string                   `json:"chat_ref,omitempty"`
	Group      string                   `json:"group"`
	Status     string                   `json:"capture_status"`
	Fixture    string                   `json:"input_fixture,omitempty"`
	Benchtime  string                   `json:"benchtime"`
	Timing     string                   `json:"timing"`
	Caveats    []string                 `json:"caveats"`
	Host       analysis.EnvironmentFact `json:"host"`
	Series     []uiSeries               `json:"series"`
	Report     string                   `json:"report"`
	Evidence   string                   `json:"evidence,omitempty"`
	ReportPath string                   `json:"report_path,omitempty"`
}

type uiSeries struct {
	Method  string     `json:"method"`
	Size    int        `json:"size"`
	Median  float64    `json:"median_ns"`
	Min     float64    `json:"min_ns"`
	Max     float64    `json:"max_ns"`
	Repeats []uiRepeat `json:"repeats"`
}

type uiRepeat struct {
	Number int     `json:"number"`
	NS     float64 `json:"ns"`
}

func newUIResult(capture evidence.Capture, inspection analysis.Interpretation, report, path, reportPath string) uiResult {
	qualifications := shortResultCaveats(inspection)
	view := uiResult{
		Group: capture.Group, Status: inspection.Status,
		Fixture: comparisonFixture(capture), Benchtime: capture.Config.Benchtime,
		Timing: comparisonTiming(inspection, qualifications != ""),
		Host:   inspection.MeasuringEnvironment, Report: report, Evidence: path, ReportPath: reportPath,
	}
	if qualifications != "" {
		view.Caveats = strings.Split(strings.TrimSpace(qualifications), "\n")
	}
	view.Series = comparisonSeries(capture, inspection)
	return view
}

func comparisonFixture(capture evidence.Capture) string {
	if len(capture.Run.Observations) == 0 {
		return ""
	}
	return capture.Run.Observations[0].Key.Case.Input.GeneratorID
}

func comparisonSeries(capture evidence.Capture, inspection analysis.Interpretation) []uiSeries {
	type seriesKey struct {
		operation string
		strategy  string
		size      int
	}
	repeats := make(map[seriesKey][]uiRepeat)
	for _, observation := range capture.Run.Observations {
		c := observation.Key.Case
		key := seriesKey{c.OperationID, c.StrategyID, c.Elements}
		repeats[key] = append(repeats[key], uiRepeat{Number: observation.Repetition, NS: observation.NanosecondsPerOp})
	}
	seriesRows := make([]uiSeries, 0, len(inspection.Series))
	for _, series := range inspection.Series {
		samples := repeats[seriesKey{series.OperationID, series.StrategyID, series.Elements}]
		slices.SortFunc(samples, func(a, b uiRepeat) int { return a.Number - b.Number })
		seriesRows = append(seriesRows, uiSeries{
			Method: seriesMethodName(series), Size: series.Elements,
			Median: series.MedianNS, Min: series.MinimumNS, Max: series.MaximumNS, Repeats: samples,
		})
	}
	return seriesRows
}
