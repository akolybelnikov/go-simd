package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		want      string
		wantError string
	}{
		{name: "default all", want: "all"},
		{name: "addition", args: []string{"-function", "add"}, want: "add"},
		{name: "compute", args: []string{"-function=compute"}, want: "compute"},
		{name: "invalid function", args: []string{"-function", "multiply"}, wantError: "unknown function"},
		{name: "unexpected argument", args: []string{"add"}, wantError: "unexpected positional arguments"},
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
			if got.function != tt.want {
				t.Errorf("function = %q, want %q", got.function, tt.want)
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
	a, b, out := makeInputs(1)
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
