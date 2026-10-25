package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"

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
		{name: "addition", args: []string{"-function", "add"}, want: "addition"},
		{name: "compute", args: []string{"-function=compute"}, want: "computation"},
		{name: "addition group", args: []string{"-group", "addition"}, want: "addition"},
		{name: "computation group", args: []string{"-group=computation"}, want: "computation"},
		{name: "addition shortcut", args: []string{"-addition"}, want: "addition"},
		{name: "computation shortcut", args: []string{"-computation"}, want: "computation"},
		{name: "invalid function", args: []string{"-function", "multiply"}, wantError: "unknown group"},
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
