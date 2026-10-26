package main

import (
	"bytes"
	"flag"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      string
		wantError string
	}{
		{name: "default all", want: "all"},
		{name: "explicit all", args: []string{"-group=all"}, want: "all"},
		{name: "disabled shortcut", args: []string{"-addition=false"}, want: "all"},
		{name: "addition", args: []string{"-function", "add"}, want: "addition"},
		{name: "compute", args: []string{"-function=compute"}, want: "computation"},
		{name: "addition group", args: []string{"-group", "addition"}, want: "addition"},
		{name: "computation group", args: []string{"-group=computation"}, want: "computation"},
		{name: "addition shortcut", args: []string{"-addition"}, want: "addition"},
		{name: "computation shortcut", args: []string{"-computation"}, want: "computation"},
		{name: "invalid function", args: []string{"-function", "multiply"}, wantError: "unknown group"},
		{name: "empty group", args: []string{"-group="}, wantError: "unknown group"},
		{name: "unexpected argument", args: []string{"add"}, wantError: "unexpected positional arguments"},
		{name: "conflicting group and shortcut", args: []string{"-group", "all", "-addition"}, wantError: "conflicting"},
		{name: "conflicting legacy", args: []string{"-function", "compute", "-group", "addition"}, wantError: "conflicting"},
		{name: "matching aliases", args: []string{"-function", "add", "-group", "addition", "-addition"}, want: "addition"},
		{name: "conflicting shortcuts", args: []string{"-addition", "-computation"}, wantError: "cannot be used together"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			got, err := parseOptions(tt.args, &stderr)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("parseOptions() error = %v, want substring %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOptions() error = %v", err)
			}
			if got.group != tt.want {
				t.Errorf("function = %q, want %q", got.group, tt.want)
			}
		})
	}
}

func TestParseOptionsHelp(t *testing.T) {
	var stderr bytes.Buffer
	_, err := parseOptions([]string{"-h"}, &stderr)
	if err != flag.ErrHelp {
		t.Fatalf("parseOptions(-h) error = %v, want %v", err, flag.ErrHelp)
	}
	if !strings.Contains(stderr.String(), "Usage: simdbench") {
		t.Errorf("help output missing usage: %q", stderr.String())
	}
	want, err := os.ReadFile("testdata/help.txt")
	if err != nil {
		t.Fatal(err)
	}
	if strings.ReplaceAll(stderr.String(), "\r\n", "\n") != string(want) {
		t.Fatalf("help differs from compatibility fixture\ngot:\n%s\nwant:\n%s", stderr.String(), want)
	}
}

func TestLegacyVariantRegistration(t *testing.T) {
	if benchmarkSize != 1<<20 {
		t.Fatalf("quick comparison size changed: %d", benchmarkSize)
	}
	a, b, out := benchutil.MakeBenchmarkSlices(1)
	groups := benchmarkGroups("all", a, b, out)
	var names [][]string
	for _, group := range groups {
		row := []string{group.name}
		for _, variant := range group.variants {
			row = append(row, variant.name)
			out[0] = 123
			if err := variant.run(); err != nil {
				t.Fatalf("%s/%s: %v", group.name, variant.name, err)
			}
			if out[0] != -2.5 && group.name == "Addition" {
				t.Fatalf("%s is no longer bound to addition", variant.name)
			}
			if out[0] != 6.25 && group.name != "Addition" {
				t.Fatalf("%s is no longer bound to computation", variant.name)
			}
		}
		names = append(names, row)
	}
	want := [][]string{
		{"Addition", "Scalar", "SIMDPartial", "SIMDTail", "SIMDParallel"},
		{"Compute-heavy expression: a*a + b*b + a*b", "Scalar", "SIMD"},
	}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("registered groups/variants = %v, want %v", names, want)
	}
}

func TestLegacyQuickComparisonTable(t *testing.T) {
	var output bytes.Buffer
	err := printResults(&output, "Addition", []result{
		{name: "Scalar", stats: testing.BenchmarkResult{N: 10, T: 100 * time.Nanosecond}},
		{name: "SIMDPartial", stats: testing.BenchmarkResult{N: 10, T: 80 * time.Nanosecond}},
		{name: "SIMDTail", stats: testing.BenchmarkResult{N: 10, T: 120 * time.Nanosecond}},
		{name: "SIMDParallel", stats: testing.BenchmarkResult{N: 10, T: 300 * time.Nanosecond, MemBytes: 1280, MemAllocs: 20}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		got = append(got, strings.Join(strings.Fields(line), " "))
	}
	want := []string{
		"Addition (1048576 float32 elements)",
		"IMPLEMENTATION TIME (ns/op) ALLOCATIONS CHANGE VS SCALAR",
		"Scalar 10.00 0 B/op, 0 allocs/op baseline",
		"SIMDPartial 8.00 0 B/op, 0 allocs/op -20.0%",
		"SIMDTail 12.00 0 B/op, 0 allocs/op +20.0%",
		"SIMDParallel 30.00 128 B/op, 2 allocs/op +200.0%",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("quick table changed\ngot: %v\nwant: %v", got, want)
	}
}

func TestBenchmarkGroups(t *testing.T) {
	a, b, out := benchutil.MakeBenchmarkSlices(1)
	for _, tt := range []struct {
		function string
		want     int
	}{
		{function: "add", want: 1},
		{function: "compute", want: 1},
		{function: "all", want: 2},
	} {
		groups := benchmarkGroups(tt.function, a, b, out)
		if len(groups) != tt.want {
			t.Errorf("benchmarkGroups(%q) returned %d groups, want %d", tt.function, len(groups), tt.want)
		}
	}
}
