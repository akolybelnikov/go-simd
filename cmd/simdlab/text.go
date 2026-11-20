package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/labapi"
)

func catalogText(response labapi.CatalogResponse) string {
	var text strings.Builder
	fmt.Fprintf(&text, "Lab host: %s/%s, %d-bit vectors, SIMD emulated: %t\n%s\n", response.Runtime.OS, response.Runtime.Arch, response.Runtime.VectorBits, response.Runtime.Emulated, response.MeasurementNotice)
	for _, operation := range response.Operations {
		fmt.Fprintf(&text, "\n%s (%s) — %s\n", operation.Group, operation.TypeID, operation.Title)
		fmt.Fprintln(&text, "  Strategy IDs:")
		for _, strategy := range operation.Strategies {
			fmt.Fprintf(&text, "    %s (%s, %s)\n", strategy.ID, strategy.Label, strategy.WorkerPolicy)
		}
		fmt.Fprintf(&text, "  Supported element counts: %v\n", operation.SupportedSizes)
	}
	return text.String()
}

func inspectionText(inspection analysis.Interpretation) string {
	var text strings.Builder
	env := inspection.MeasuringEnvironment
	fmt.Fprintf(&text, "Run %s (%s)\nMeasuring host: %s/%s, %d-bit vectors, SIMD emulated: %t\n%s\n", inspection.RunID, inspection.Status, env.OS, env.Arch, env.VectorBits, env.Emulated, inspection.MeasurementNotice)
	fmt.Fprintln(&text, "\nObserved kernel timings:")
	for _, series := range inspection.Series {
		fmt.Fprintf(&text, "  %s at %d elements: median %.6g ns/op (%d observations, range %.6g to %.6g ns/op)\n", series.StrategyID, series.Elements, series.MedianNS, series.Samples, series.MinimumNS, series.MaximumNS)
	}
	if len(inspection.Comparisons) > 0 {
		fmt.Fprintln(&text, "\nComparisons (candidate vs. scalar baseline):")
		for _, comparison := range inspection.Comparisons {
			fmt.Fprintf(&text, "  %s at %d elements: %+.6g%% median time change, %.6gx baseline/candidate\n", comparison.CandidateStrategy, comparison.Elements, comparison.TimeChangePercent, comparison.SpeedupRatio)
		}
	}
	fmt.Fprintln(&text, "\nCaveats:")
	for _, caveat := range inspection.Caveats {
		fmt.Fprintf(&text, "  - %s\n", caveat.Message)
	}
	return text.String()
}

func writeText(stdout, stderr io.Writer, text string) int {
	if _, err := io.WriteString(stdout, terminalText(text)); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
