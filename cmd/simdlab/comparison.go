package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/labtools"
	"github.com/akolybelnikov/go-simd/internal/labworker"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

const starterSize = 1024

func comparisonRequest(group string, variants []string, size int) labapi.PlanRequest {
	return labapi.PlanRequest{Group: group, Variants: variants, Sizes: []int{size}, Count: 3, Benchtime: "100ms"}
}

func additionStarter() labapi.PlanRequest {
	return comparisonRequest("addition", []string{catalog.Scalar, catalog.SIMDPartial}, starterSize)
}

func runStarter(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("artifacts", filepath.Join("artifacts", "experiments"), "local artifact directory")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: simdlab run [-artifacts directory] addition")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 1 || flags.Arg(0) != "addition" || *root == "" {
		flags.Usage()
		return 2
	}
	tools, err := newComparisonTools(*root)
	if err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return 1
	}
	plan, err := tools.ProposeExperiment(additionStarter())
	if err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return 1
	}
	if writeText(stdout, stderr, comparisonPlanText(plan)) != 0 {
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	_, err = measureComparison(ctx, tools, plan, stdout, stderr)
	if err != nil {
		fmt.Fprintln(stderr, terminalText(err.Error()))
		return 1
	}
	return 0
}

func newComparisonTools(root string) (*labtools.Toolset, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	runner, err := labworker.NewRunner(executable, "_worker")
	if err != nil {
		return nil, err
	}
	return labtools.New(labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy(), experimentExecutor{runner: runner, root: absolute})
}

func comparisonPlanText(plan labapi.ExperimentPlan) string {
	var text strings.Builder
	fmt.Fprintf(&text, "Compare %s on %s/%s.\n", plan.Group, plan.Runtime.OS, plan.Runtime.Arch)
	for _, c := range plan.Cases {
		fmt.Fprintf(&text, "  %s · %d values per input array\n", methodName(plan.Group, c.StrategyID, c.StrategyLabel), c.Elements)
	}
	fmt.Fprintf(&text, "%d repeats per method · %s target per repeat · generated %s inputs\n", plan.Count, plan.Benchtime, plan.TypeID)
	for _, warning := range plan.Warnings {
		fmt.Fprintln(&text, warning.Message)
	}
	return text.String()
}

func methodName(group, id, label string) string {
	if group == "addition" {
		switch id {
		case catalog.Scalar:
			return "Ordinary addition"
		case catalog.SIMDPartial:
			return "Vectorized addition (partial vectors)"
		}
	}
	if group == "sum" {
		switch id {
		case catalog.Scalar:
			return "Ordinary sum (left to right)"
		case catalog.SIMDStriped:
			return "Vectorized sum (one accumulator)"
		case catalog.SIMDStriped2:
			return "Vectorized sum (two accumulators)"
		case catalog.SIMDStriped4:
			return "Vectorized sum (four accumulators)"
		}
	}
	return label
}

type savedComparison struct {
	inspection analysis.Interpretation
	capture    *evidence.Capture
	path       string
	reportPath string
	report     string
}

func measureComparison(ctx context.Context, tools *labtools.Toolset, plan labapi.ExperimentPlan, stdout, stderr io.Writer) (savedComparison, error) {
	if _, err := fmt.Fprintln(stdout, "Measuring… Press Ctrl+C to cancel."); err != nil {
		return savedComparison{}, err
	}
	saved, err := executeComparison(ctx, tools, plan)
	if err != nil {
		if saved.capture != nil && saved.inspection.Status == "" {
			fmt.Fprintln(stderr, terminalText("Partial measurements: "+saved.path))
			return savedComparison{}, err
		}
		return saved, err
	}
	summary := comparisonSummary(saved.inspection) + "This run used generated data on this machine.\n"
	if _, err := io.WriteString(stdout, terminalText(summary+"Report: "+saved.reportPath+"\nMeasurements: "+saved.path+"\n")); err != nil {
		return saved, err
	}
	return saved, nil
}

func executeComparison(ctx context.Context, tools *labtools.Toolset, plan labapi.ExperimentPlan) (savedComparison, error) {
	if err := ctx.Err(); err != nil {
		return savedComparison{}, err
	}
	if err := tools.ConfirmExperiment(plan.ID); err != nil {
		return savedComparison{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, experimentRunTimeout)
	defer cancel()
	run, err := tools.ExecuteConfirmed(ctx, plan.ID)
	path := filepath.Join(run.Directory, "measurements.json")
	if err != nil {
		var failed savedComparison
		if run.Capture != nil {
			failed = savedComparison{capture: run.Capture, path: path}
		}
		if ctx.Err() != nil {
			return failed, fmt.Errorf("measurement interrupted: %w", ctx.Err())
		}
		return failed, err
	}
	if run.Capture == nil {
		return savedComparison{}, errors.New("measurement completed without saved evidence")
	}
	inspection, err := analysis.InterpretCapture(*run.Capture)
	if err != nil {
		return savedComparison{}, fmt.Errorf("read results in %s: %w", path, err)
	}
	reportPath := filepath.Join(run.Directory, "report.md")
	report := comparisonReport(plan.Group, inspection, true)
	saved := savedComparison{inspection: inspection, capture: run.Capture, path: path}
	if err := os.WriteFile(reportPath, []byte(report), 0o600); err != nil {
		return saved, fmt.Errorf("write report (measurements saved at %s): %w", path, err)
	}
	saved.reportPath, saved.report = reportPath, report
	return saved, nil
}

func comparisonReport(group string, inspection analysis.Interpretation, local bool) string {
	summary := comparisonSummary(inspection)
	if local {
		summary += "This run used generated data on this machine.\n\nMeasurements: [measurements.json](measurements.json)\n"
	}
	details := strings.TrimSuffix(strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(inspectionText(inspection)), "\n")
	return "# " + group + " comparison\n\n" + summary + "\n## Measurement details\n\n    " +
		strings.ReplaceAll(details, "\n", "\n    ") + "\n"
}

func comparisonSummary(inspection analysis.Interpretation) string {
	qualifications := shortResultCaveats(inspection)
	return comparisonTiming(inspection, qualifications != "") + qualifications
}

func comparisonTiming(inspection analysis.Interpretation, limited bool) string {
	var text strings.Builder
	for _, c := range inspection.Comparisons {
		baseline := findSeries(inspection, c.BaselineStrategy, c.Elements)
		candidate := findSeries(inspection, c.CandidateStrategy, c.Elements)
		first, second := candidate, baseline
		if c.TimeChangePercent > 0 {
			first, second = baseline, candidate
		}
		firstName := seriesMethodName(first)
		secondName := seriesMethodName(second)
		if c.TimeChangePercent == 0 {
			fmt.Fprintf(&text, "%s and %s had equal median times for the %d-value batch.\n", firstName, secondName, c.Elements)
		} else {
			reduction := (second.MedianNS - first.MedianNS) / second.MedianNS * 100
			percent := fmt.Sprintf("%.3g%%", reduction)
			if percent == "100%" {
				percent = "over 99.9%"
			}
			if limited {
				fmt.Fprintf(&text, "%s had a %s lower median in this run than %s for the %d-value batch.\n", firstName, percent, secondName, c.Elements)
			} else {
				fmt.Fprintf(&text, "%s took %s less time than %s for the %d-value batch.\n", firstName, percent, secondName, c.Elements)
			}
		}
		fmt.Fprintf(&text, "Median per batch: %.4g ns for %s and %.4g ns for %s", first.MedianNS, firstName, second.MedianNS, secondName)
		if first.Samples == second.Samples {
			fmt.Fprintf(&text, " across %d repeats per method.\n", first.Samples)
		} else {
			fmt.Fprintf(&text, " (%d and %d repeats, respectively).\n", first.Samples, second.Samples)
		}
	}
	if text.Len() == 0 {
		fmt.Fprintln(&text, "The saved run has no compatible timing comparison.")
	}
	return text.String()
}

func shortResultCaveats(inspection analysis.Interpretation) string {
	var text strings.Builder
	if inspection.Status != "complete" {
		fmt.Fprintln(&text, "This run is incomplete. The results cover the repeats saved before it stopped.")
	}
	if inspection.MeasuringEnvironment.Emulated {
		fmt.Fprintln(&text, "This run used SIMD emulation.")
	}
	for _, series := range inspection.Series {
		if series.Samples < 3 {
			fmt.Fprintln(&text, "At least one method has fewer than three repeats, so timing variability is uncertain.")
			break
		}
	}
	for _, caveat := range inspection.Caveats {
		switch caveat.Code {
		case "fixed_iterations":
			fmt.Fprintln(&text, "This run used a fixed iteration count. Short measurements can make timings less stable.")
		case "different_fold_orders":
			fmt.Fprintln(&text, "The methods add values in different orders, which can change the rounded result.")
		case "parallel_vs_serial":
			fmt.Fprintln(&text, "The comparison includes differences in concurrency as well as implementation.")
		}
	}
	return text.String()
}

func findSeries(inspection analysis.Interpretation, strategy string, elements int) analysis.SeriesFact {
	for _, series := range inspection.Series {
		if series.StrategyID == strategy && series.Elements == elements {
			return series
		}
	}
	return analysis.SeriesFact{}
}

func seriesMethodName(series analysis.SeriesFact) string {
	for _, operation := range catalog.All() {
		if operation.ID == series.OperationID {
			return methodName(operation.Group, series.StrategyID, series.Strategy)
		}
	}
	return series.Strategy
}

func repeatSummary(inspection analysis.Interpretation) string {
	var text strings.Builder
	for _, series := range inspection.Series {
		fmt.Fprintf(&text, "%s at %d values: %d repeats, %.4g–%.4g ns per batch (median %.4g ns).\n",
			seriesMethodName(series), series.Elements, series.Samples, series.MinimumNS, series.MaximumNS, series.MedianNS)
	}
	return text.String()
}
