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

	"github.com/akolybelnikov/go-simd/packages/ops"
)

const benchmarkSize = 1 << 20

var benchmarkSink float32

type options struct {
	function string
}

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

	a, b, out := makeInputs(benchmarkSize)
	for _, group := range benchmarkGroups(opts.function, a, b, out) {
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
	flags := flag.NewFlagSet("simdbench", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&opts.function, "function", "all", "function to benchmark: add, compute, or all")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: simdbench [-function add|compute|all]")
		fmt.Fprintln(stderr, "\nBenchmarks scalar and SIMD operations on a fixed 1,048,576-element input.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	switch opts.function {
	case "add", "compute", "all":
		return opts, nil
	default:
		return options{}, fmt.Errorf("unknown function %q (choose add, compute, or all)", opts.function)
	}
}

func benchmarkGroups(function string, a, b, out []float32) []benchmarkGroup {
	add := benchmarkGroup{
		name: "Addition",
		variants: []variant{
			{name: "Scalar", run: func() error { return ops.AddScalar(a, b, out) }},
			{name: "SIMD partial", run: func() error { return ops.AddSIMD(a, b, out) }},
			{name: "SIMD tail", run: func() error { return ops.AddSIMDVectorTail(a, b, out) }},
			{name: "SIMD parallel", run: func() error { return ops.AddSIMDParallel(a, b, out) }},
		},
	}
	compute := benchmarkGroup{
		name: "Compute-heavy expression: a*a + b*b + a*b",
		variants: []variant{
			{name: "Scalar", run: func() error { return ops.ComputeHeavyScalar(a, b, out) }},
			{name: "SIMD", run: func() error { return ops.ComputeHeavySIMD(a, b, out) }},
		},
	}

	switch function {
	case "add":
		return []benchmarkGroup{add}
	case "compute":
		return []benchmarkGroup{compute}
	default:
		return []benchmarkGroup{add, compute}
	}
}

func makeInputs(size int) (a, b, out []float32) {
	a = make([]float32, size)
	b = make([]float32, size)
	out = make([]float32, size)
	for i := range a {
		a[i] = float32(i%100) / 10
		b[i] = float32(i%50-25) / 10
	}
	return a, b, out
}

func printResults(w io.Writer, title string, results []result) error {
	fmt.Fprintf(w, "\n%s (%d float32 elements)\n", title, benchmarkSize)
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
