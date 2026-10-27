package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"simd"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/cli"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/akolybelnikov/go-simd/internal/report"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "simdreport:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("simdreport", flag.ContinueOnError)
	flags.Usage = func() {
		cli.Help(flags.Output(), flags, "Repeated kernel measurements and TXT/MD/CSV reports. Runs standalone by default.", "simdreport -list", "simdreport -group addition -bench '^BenchmarkAddition$/^n=1024$' -count 3 -benchtime 100ms -format md", "simdreport -backend go-test -group computation")
	}
	var selection cli.Selection
	selection.Register(flags)
	backend := flags.String("backend", "compiled", "measurement backend: compiled or go-test (developer checkout only)")
	count := flags.Int("count", 5, "measurement repetitions")
	benchtime := flags.String("benchtime", "100ms", "duration or iteration count per measurement")
	bench := flags.String("bench", ".", "Go benchmark selection expression")
	output := flags.String("output", "artifacts", "report directory")
	formatName := flags.String("format", "txt", "saved report format: txt, md, or csv (terminal stays text)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	group, err := selection.Resolve(flags)
	if err != nil {
		return err
	}
	config := engine.Config{Count: *count, Benchtime: *benchtime, Bench: *bench}
	if err := config.Validate(); err != nil {
		return err
	}
	format, err := report.ParseFormat(*formatName)
	if err != nil {
		return err
	}
	if *backend != "compiled" && *backend != "go-test" {
		return fmt.Errorf("unknown backend %q (choose compiled or go-test)", *backend)
	}
	machine := platform.Current()
	if selection.List {
		return cli.PrintCatalog(stdout, group, machine)
	}
	var groups []string
	for _, operation := range catalog.All() {
		if !slices.Contains(groups, operation.Group) {
			groups = append(groups, operation.Group)
		}
	}
	if group != "all" {
		groups = []string{group}
	}
	plans := make(map[string]engine.Plan)
	for _, name := range groups {
		plan, err := engine.Prepare(name, config, machine.Lanes)
		if err != nil {
			return err
		}
		plans[name] = plan
	}
	for _, name := range groups {
		var err error
		if *backend == "go-test" {
			err = measureLegacyReport(name, *bench, *benchtime, *count, *output, format, stdout)
		} else {
			err = measureCompiledReport(name, plans[name], machine, *output, format, stdout)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func measureLegacyReport(group, bench, benchtime string, count int, output string, format report.Format, stdout io.Writer) error {
	directory := filepath.Join(output, group)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	raw, err := os.CreateTemp(directory, "benchmark-"+time.Now().UTC().Format("20060102-150405")+"-*.raw")
	if err != nil {
		return err
	}
	args := []string{"test", "./internal/" + group, "-run=^$", "-bench=" + bench, "-benchmem", "-benchtime=" + benchtime, "-count=" + strconv.Itoa(count)}
	var vector simd.Float32s
	metadata := fmt.Sprintf("date: %s\ngo: %s\nplatform: %s/%s\nGOMAXPROCS: %d\nSIMD float32 lanes: %d\ncommand: go %s\n", time.Now().UTC().Format(time.RFC3339), runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), vector.Len(), strings.Join(args, " "))
	envOutput, err := exec.Command("go", "env", "-json", "GOEXPERIMENT", "GOAMD64", "GOARM64", "GOTOOLCHAIN").CombinedOutput()
	if err != nil {
		raw.Close()
		return fmt.Errorf("go env: %w: %s", err, envOutput)
	}
	metadata += "toolchain environment: " + string(envOutput) + "\n"
	if _, err := io.WriteString(raw, metadata); err != nil {
		raw.Close()
		return err
	}
	fmt.Fprintf(stdout, "Measuring %s (%d repetitions)...\n", group, count)
	var measurements bytes.Buffer
	command := exec.Command("go", args...)
	command.Stdout = io.MultiWriter(raw, &measurements)
	command.Stderr = command.Stdout
	runErr := command.Run()
	closeErr := raw.Close()
	if runErr != nil {
		return fmt.Errorf("benchmark failed (raw output saved to %s): %w", raw.Name(), runErr)
	}
	if closeErr != nil {
		return closeErr
	}
	// Preserve the benchmark runner's CPU identification in the readable report too.
	for _, line := range strings.Split(measurements.String(), "\n") {
		if strings.HasPrefix(line, "cpu:") {
			metadata += strings.TrimSpace(line) + "\n"
			break
		}
	}
	run, err := result.ReadLegacy(&measurements, result.LegacyOptions{
		RunID: strings.TrimSuffix(filepath.Base(raw.Name()), ".raw"), Metadata: metadata, GeneratedInputs: true,
	})
	if err != nil {
		return fmt.Errorf("%s (raw output saved to %s)", err, raw.Name())
	}
	return saveReport(group, raw.Name(), run, format, stdout)
}

func saveReport(group, evidencePath string, run result.Run, format report.Format, stdout io.Writer) error {
	summaries, err := result.Summarize(run.Observations)
	if err != nil {
		return fmt.Errorf("%w (measurements saved to %s)", err, evidencePath)
	}
	table, err := result.Compare(group, summaries)
	if err != nil {
		return fmt.Errorf("%w (measurements saved to %s)", err, evidencePath)
	}
	var saved, terminal bytes.Buffer
	if err := report.Write(&saved, format, table, run.Metadata); err != nil {
		return err
	}
	if err := report.Write(&terminal, report.TXT, table, ""); err != nil {
		return err
	}
	base := strings.TrimSuffix(strings.TrimSuffix(evidencePath, ".raw"), ".measurements.json")
	reportPath := base + "." + string(format)
	if err := os.WriteFile(reportPath, saved.Bytes(), 0644); err != nil {
		return fmt.Errorf("saving report (measurements saved to %s): %w", evidencePath, err)
	}
	_, err = fmt.Fprintf(stdout, "%s\nSaved %s and %s\n", terminal.String(), reportPath, evidencePath)
	return err
}
