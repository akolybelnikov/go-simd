package main

import (
	"slices"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/evidence"
)

type uiResult struct {
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
	Repeats []uiRepeat `json:"repeats"`
}

type uiRepeat struct {
	Number int     `json:"number"`
	NS     float64 `json:"ns"`
}

func newUIResult(capture evidence.Capture, inspection analysis.Interpretation, report, path, reportPath string) uiResult {
	qualifications := shortResultCaveats(inspection)
	view := uiResult{
		Timing: comparisonTiming(inspection, qualifications != ""),
		Host:   inspection.MeasuringEnvironment, Report: report, Evidence: path, ReportPath: reportPath,
	}
	if qualifications != "" {
		view.Caveats = strings.Split(strings.TrimSpace(qualifications), "\n")
	}
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
	for _, series := range inspection.Series {
		samples := repeats[seriesKey{series.OperationID, series.StrategyID, series.Elements}]
		slices.SortFunc(samples, func(a, b uiRepeat) int { return a.Number - b.Number })
		view.Series = append(view.Series, uiSeries{
			Method: seriesMethodName(series), Size: series.Elements, Repeats: samples,
		})
	}
	return view
}
