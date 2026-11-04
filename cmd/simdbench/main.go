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
var benchmarkByteSink []uint8

type options struct {
	input input.Options
	group string
	list  bool
}

type variant struct {
	reduceChecked func() (float32, error)
	reduce        func() float32
	name          string
	run           func() error
}

type result struct {
	name  string
	stats testing.BenchmarkResult
}

type benchmarkGroup struct {
	typeID        string
	float64Output []float64
	copyOutput    bool
	byteOutput    []uint8
	scalar        *float32
	output        []float32
	name          string
	logicalBytes  int64
	variants      []variant
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
			var stats testing.BenchmarkResult
			if item.reduceChecked != nil {
				stats = testing.Benchmark(func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(group.logicalBytes)
					b.ResetTimer()
					if runErr != nil {
						b.StopTimer()
						return
					}
					var value float32
					for i := 0; i < b.N; i++ {
						value, runErr = item.reduceChecked()
						if runErr != nil {
							b.StopTimer()
							return
						}
					}
					b.StopTimer()
					*group.scalar = value
					benchmarkSink = value
				})
			} else if item.reduce != nil {
				stats = testing.Benchmark(func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(group.logicalBytes)
					b.ResetTimer()
					var value float32
					for i := 0; i < b.N; i++ {
						value = item.reduce()
					}
					b.StopTimer()
					*group.scalar = value
					benchmarkSink = value
				})
			} else if group.byteOutput != nil {
				stats = testing.Benchmark(func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(group.logicalBytes)
					b.ResetTimer()
					if runErr != nil {
						b.StopTimer()
						return
					}
					for i := 0; i < b.N; i++ {
						if runErr = item.run(); runErr != nil {
							b.StopTimer()
							return
						}
					}
					b.StopTimer()
					benchmarkByteSink = group.byteOutput
				})
			} else if group.float64Output != nil {
				stats, runErr = measureQuickFloat64(item.run, group.float64Output, group.logicalBytes)
			} else if group.copyOutput {
				stats, runErr = measureQuickCopy(item.run, group.output, group.logicalBytes)
			} else {
				stats = testing.Benchmark(func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(group.logicalBytes)
					for i := 0; i < b.N; i++ {
						if runErr = item.run(); runErr != nil {
							b.Fatal(runErr)
						}
					}
					benchmarkSink = group.output[len(group.output)-1]
				})
			}
			if runErr != nil {
				return fmt.Errorf("%s %s benchmark: %w", group.name, item.name, runErr)
			}
			results = append(results, result{name: item.name, stats: stats})
		}
		if err := printTypedResults(stdout, group.name, group.typeID, results); err != nil {
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
		cli.Help(stderr, flags, "Quick comparison on 1,048,576 elements; -list inspects without measuring.", "simdbench -list", "simdbench -group addition", "simdbench -group computation")
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
		var third []float32
		var doubleA, doubleB, doubleOut []float64
		var byteA, byteB, byteOut []uint8
		if descriptor.ID == catalog.SaturatingAdd8 {
			byteA, byteB, byteOut = input.MakeSaturatingAdd8(size)
			resultOut = nil
		} else if descriptor.ID == catalog.Add64 {
			doubleA, doubleB, doubleOut = input.MakeAddition64(size)
			resultOut = nil
		} else if descriptor.ID == catalog.MulAdd {
			x, y, third, resultOut = input.MakeMulAdd(size, options)
		} else if descriptor.ID == catalog.Copy || descriptor.ID == catalog.BitRoundTrip || descriptor.Parameterized() || descriptor.ScalarResult() {
			x, y, resultOut = input.MakeForOperation(descriptor, size, options)
		} else if out == nil {
			a, b, out = benchutil.MakeBenchmarkSlices(size)
			x, y, resultOut = a, b, out
		}
		item := benchmarkGroup{name: descriptor.Title, logicalBytes: bytes, output: resultOut, typeID: descriptor.TypeID, float64Output: doubleOut, byteOutput: byteOut}
		if descriptor.ID == catalog.Add64 {
			item.name += "\nfixture: " + options.Generator(descriptor) + "; actual float64 lanes; logical traffic 24*n"
		}
		if descriptor.ID == catalog.Copy {
			item.copyOutput = true
			item.name += "\nfixture: " + options.Generator(descriptor) + "; disjoint buffers; GoCopy uses Go compiler/runtime copy"
		}
		if descriptor.ID == catalog.Clamp || descriptor.ID == catalog.MulAdd || descriptor.ID == catalog.BitRoundTrip {
			item.copyOutput = true
		}
		if descriptor.ByteResult() && item.byteOutput == nil {
			length, err := descriptor.OutputElements(size)
			if err != nil {
				return nil, err
			}
			item.byteOutput = make([]uint8, length)
		}
		if descriptor.ScalarResult() {
			item.scalar = new(float32)
			item.name += "\nfixture: " + options.Generator(descriptor)
		}
		if descriptor.Parameterized() {
			item.name += "\n" + cli.InputSummary(options.Identity(descriptor), options.Generator(descriptor))
		}
		if descriptor.ID == catalog.AffineFMA || descriptor.ID == catalog.MulAdd {
			item.name += "\nFMA32 SIMD path: " + platform.Current().FMA32Path
		}
		if descriptor.ID == catalog.DotFMA {
			item.name += "\nFMA32 vector gate: " + platform.Current().FMA32Path + "; scalar tails use software FMA32"
		}
		if descriptor.ID == catalog.MulAdd {
			item.name += "\nfixture: " + options.Generator(descriptor) + "; three streamed arrays, logical traffic 16*n"
		}
		if descriptor.ID == catalog.BitRoundTrip {
			item.name += "\nfixture: " + options.Generator(descriptor) + "; exact bits through uint32 reinterpretation, not value conversion; logical traffic 8*n"
		}
		if descriptor.ID == catalog.SaturatingAdd8 {
			item.name += "\nfixture: " + options.Generator(descriptor) + "; true uint8 arrays, saturation at 255; logical traffic 3*n"
		}
		for _, implementation := range operation.Implementations {
			if implementation.BindByteArrays != nil {
				item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, run: implementation.BindByteArrays(byteA, byteB, item.byteOutput)})
				continue
			}
			if implementation.BindFloat64 != nil {
				item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, run: implementation.BindFloat64(doubleA, doubleB, doubleOut)})
				continue
			}
			if implementation.BindTernary != nil {
				item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, run: implementation.BindTernary(x, y, third, resultOut)})
				continue
			}
			if implementation.BindUnary != nil {
				item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, run: implementation.BindUnary(x, resultOut)})
				continue
			}
			if descriptor.ByteResult() {
				item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, run: implementation.BindUint8(x, item.byteOutput, options)})
				continue
			}
			if implementation.BindCheckedScalar != nil {
				item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, reduceChecked: implementation.BindCheckedScalar(x, y)})
				continue
			}
			if descriptor.ScalarResult() {
				item.variants = append(item.variants, variant{name: implementation.Descriptor.Label, reduce: implementation.BindScalar(x)})
				continue
			}
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
	return printTypedResults(w, title, catalog.Float32, results)
}
func printTypedResults(w io.Writer, title, elementType string, results []result) error {
	if elementType == "" {
		elementType = catalog.Float32
	}
	if _, err := fmt.Fprintf(w, "\n%s (%d %s elements)\n", title, benchmarkSize, elementType); err != nil {
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
