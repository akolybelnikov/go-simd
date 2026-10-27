package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"testing"
	"text/tabwriter"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/cli"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

const benchmarkSize = 1 << 20

var benchmarkSink float32

type options struct {
	input input.Options
	group string
	list  bool
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
	output       []float32
	name         string
	logicalBytes int64
	variants     []variant
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
	if opts.list {
		return cli.PrintCatalogWithOptions(stdout, opts.group, platform.Current(), opts.input)
	}

	groups, err := benchmarkGroupsForSize(opts.group, benchmarkSize, nil, nil, nil, opts.input)
	if err != nil {
		return err
	}
	for _, group := range groups {
		results := make([]result, 0, len(group.variants))
		for _, item := range group.variants {
			var runErr error
			stats := testing.Benchmark(func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(group.logicalBytes)
				for i := 0; i < b.N; i++ {
					if runErr = item.run(); runErr != nil {
						b.Fatal(runErr)
					}
				}
				benchmarkSink = group.output[len(group.output)-1]
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
	var selection cli.Selection
	flags := flag.NewFlagSet("simdbench", flag.ContinueOnError)
	flags.SetOutput(stderr)
	selection.Register(flags)
	flags.Usage = func() {
		cli.Help(stderr, flags, "Quick comparison on 1,048,576 float32 elements; -list inspects without measuring.", "simdbench -list", "simdbench -group addition", "simdbench -group computation")
	}
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	group, err := selection.Resolve(flags)
	if err != nil {
		return options{}, err
	}
	opts.group, opts.list = group, selection.List
	opts.input = selection.Input
	return opts, nil
}

func benchmarkGroups(group string, a, b, out []float32) ([]benchmarkGroup, error) {
	return benchmarkGroupsForSize(group, len(a), a, b, out, input.DefaultOptions())
}

func benchmarkGroupsForSize(group string, size int, a, b, out []float32, options input.Options) ([]benchmarkGroup, error) {
	switch group {
	case "add":
		group = "addition"
	case "compute":
		group = "computation"
	}
	var groups []benchmarkGroup
	for _, descriptor := range catalog.All() {
		if group == "all" || group == descriptor.Group {
			if err := options.ValidateFor(descriptor); err != nil {
				return nil, err
			}
		}
	}
	for _, descriptor := range catalog.All() {
		if group != "all" && group != descriptor.Group {
			continue
		}
		operation, err := kernels.Lookup(descriptor.ID, descriptor.TypeID)
		if err != nil {
			return nil, err
		}
		bytes, err := descriptor.LogicalBytes(size)
		if err != nil {
			return nil, err
		}
		x, y, resultOut := a, b, out
		if descriptor.Parameterized() {
			x, y, resultOut = input.MakeForOperation(descriptor, size, options)
		} else if out == nil {
			a, b, out = benchutil.MakeBenchmarkSlices(size)
			x, y, resultOut = a, b, out
		}
		item := benchmarkGroup{name: descriptor.Title, logicalBytes: bytes, output: resultOut}
		if descriptor.Parameterized() {
			item.name += "\n" + cli.InputSummary(options.Identity(descriptor), options.Generator(descriptor))
		}
		for _, implementation := range operation.Implementations {
			var call func() error
			if descriptor.Parameterized() {
				call = implementation.BindConfigured(x, y, resultOut, options)
			} else {
				call = implementation.Bind(x, y, resultOut)
			}
			item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, run: call})
		}
		groups = append(groups, item)
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("unknown group %q", group)
	}
	return groups, nil
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
