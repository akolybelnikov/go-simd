package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/computation"
)

const benchmarkSize = 1 << 20

var benchmarkSink float32

type options struct{ group string }

type variant struct {
	name string
	run  func() error
}

type result struct {
	name  string
	stats testing.BenchmarkResult
}

type benchmarkGroup struct {
	name     string
	variants []variant
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "simdbench:", err)
		os.Exit(2)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	opts, err := parseOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	a, b, out := benchutil.MakeBenchmarkSlices(benchmarkSize)
	for _, group := range benchmarkGroups(opts.group, a, b, out) {
		results := make([]result, 0, len(group.variants))
		for _, item := range group.variants {
			var runErr error
			stats := testing.Benchmark(func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(benchmarkSize * 3 * 4))
				for i := 0; i < b.N; i++ {
					if runErr = item.run(); runErr != nil {
						b.Fatal(runErr)
					}
				}
				benchmarkSink = out[len(out)-1]
			})
			if runErr != nil {
				return fmt.Errorf("%s %s benchmark: %w", group.name, item.name, runErr)
			}
			results = append(results, result{name: item.name, stats: stats})
		}
		if err := printResults(stdout, group.name, results); err != nil {
			return err
		}
	}
	return nil
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	opts := options{}
	var legacy, group string
	var add, compute bool
	flags := flag.NewFlagSet("simdbench", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&legacy, "function", "", "deprecated alias: add, compute, or all")
	flags.StringVar(&group, "group", "", "benchmark group: addition, computation, or all")
	flags.BoolVar(&add, "addition", false, "alias for -group addition")
	flags.BoolVar(&compute, "computation", false, "alias for -group computation")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: simdbench [-group addition|computation|all] [-addition] [-computation]")
		fmt.Fprintln(stderr, "\nBenchmarks scalar and SIMD operations on a fixed 1,048,576-element input.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	selections := []string{}
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "function":
			selections = append(selections, legacy)
		case "group":
			selections = append(selections, group)
		case "addition":
			if add {
				selections = append(selections, "addition")
			}
		case "computation":
			if compute {
				selections = append(selections, "computation")
			}
		}
	})
	opts.group = "all"
	for i, selection := range selections {
		switch selection {
		case "add":
			selection = "addition"
		case "compute":
			selection = "computation"
		case "addition", "computation", "all":
		default:
			return options{}, fmt.Errorf("unknown group %q (choose addition, computation, or all)", selection)
		}
		if i > 0 && selection != opts.group {
			return options{}, fmt.Errorf("conflicting group selections cannot be used together")
		}
		opts.group = selection
	}
	return opts, nil
}

func benchmarkGroups(function string, a, b, out []float32) []benchmarkGroup {
	add := benchmarkGroup{
		name: "Addition",
		variants: []variant{
			{name: "Scalar", run: func() error { return addition.AddScalar(a, b, out) }},
			{name: "SIMDPartial", run: func() error { return addition.AddSIMD(a, b, out) }},
			{name: "SIMDTail", run: func() error { return addition.AddSIMDVectorTail(a, b, out) }},
			{name: "SIMDParallel", run: func() error { return addition.AddSIMDParallel(a, b, out) }},
		},
	}
	compute := benchmarkGroup{
		name: "Compute-heavy expression: a*a + b*b + a*b",
		variants: []variant{
			{name: "Scalar", run: func() error { return computation.ComputeHeavyScalar(a, b, out) }},
			{name: "SIMD", run: func() error { return computation.ComputeHeavySIMD(a, b, out) }},
		},
	}

	switch function {
	case "add", "addition":
		return []benchmarkGroup{add}
	case "compute", "computation":
		return []benchmarkGroup{compute}
	default:
		return []benchmarkGroup{add, compute}
	}
}

func printResults(w io.Writer, title string, results []result) error {
	if _, err := fmt.Fprintf(w, "\n%s (%d float32 elements)\n", title, benchmarkSize); err != nil {
		return err
	}
	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "IMPLEMENTATION\tTIME (ns/op)\tALLOCATIONS\tCHANGE VS SCALAR")
	var scalarNanos float64
	for _, item := range results {
		if item.name == "Scalar" {
			scalarNanos = float64(item.stats.NsPerOp())
		}
	}
	for _, item := range results {
		nanos := float64(item.stats.NsPerOp())
		change := "baseline"
		if item.name != "Scalar" && scalarNanos > 0 {
			change = fmt.Sprintf("%+.1f%%", (nanos-scalarNanos)/scalarNanos*100)
		}
		fmt.Fprintf(
			table,
			"%s\t%.2f\t%d B/op, %d allocs/op\t%s\n",
			item.name,
			float64(item.stats.NsPerOp()),
			item.stats.AllocedBytesPerOp(),
			item.stats.AllocsPerOp(),
			change,
		)
	}
	return table.Flush()
}
