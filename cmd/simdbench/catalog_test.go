package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestListSelectors(t *testing.T) {
	for _, tt := range []struct {
		args  []string
		group string
		list  bool
	}{
		{[]string{"-list"}, "all", true},
		{[]string{"-list=false"}, "all", false},
		{[]string{"-list", "-function=add"}, "addition", true},
		{[]string{"-list=true", "-computation"}, "computation", true},
	} {
		got, err := parseOptions(tt.args, io.Discard)
		if err != nil || got.group != tt.group || got.list != tt.list {
			t.Fatalf("parseOptions(%v) = %+v, %v", tt.args, got, err)
		}
	}
	for _, args := range [][]string{{"-list", "-group=broadcast"}, {"-list", "-addition", "-computation"}, {"-list", "extra"}} {
		if err := run(args, io.Discard, io.Discard); err == nil {
			t.Fatalf("invalid listing selection accepted: %v", args)
		}
	}
}

func TestCatalogListing(t *testing.T) {
	for _, emulated := range []bool{false, true} {
		var output bytes.Buffer
		err := printCatalog(&output, "all", capabilities{"go-fixture", "fixture-os", "fixture-arch", 256, 8, 4, emulated})
		if err != nil {
			t.Fatal(err)
		}
		text := output.String()
		for _, want := range []string{
			"no benchmarks run", "go: go-fixture", "platform: fixture-os/fixture-arch", "SIMD bits: 256", "SIMD float32 lanes: 8", "GOMAXPROCS: 4",
			"Addition [add/float32]", "[compute-expression/float32]", "go-float32-add-v1", "go-float32-quadratic-v1",
			"inputs: a:float32[], b:float32[]", "output: out:float32[]", "scalar parameters: 0", "logical traffic at n=1: 12 bytes/call",
			"scalar reference: scalar", "execution: kernel-only", "per-call-gomaxprocs", "simd-parallel-scalar-tail",
		} {
			if !strings.Contains(text, want) {
				t.Errorf("listing missing %q:\n%s", want, text)
			}
		}
		mode := "portable hardware"
		if emulated {
			mode = "portable emulation"
		}
		if strings.Count(text, mode) != 4 {
			t.Errorf("expected four SIMD variants in %q mode:\n%s", mode, text)
		}
		if strings.Contains(text, "ns/op") || strings.Contains(text, "broadcast/") {
			t.Fatal("listing contains measurements or planned kernels")
		}
	}
}

func TestRunListOnly(t *testing.T) {
	for _, tt := range []struct{ group, present, absent string }{
		{"addition", "Addition [add/float32]", "[compute-expression/float32]"},
		{"computation", "[compute-expression/float32]", "Addition [add/float32]"},
	} {
		var output, stderr bytes.Buffer
		if err := run([]string{"-list", "-group=" + tt.group}, &output, &stderr); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), tt.present) || strings.Contains(output.String(), tt.absent) || strings.Contains(output.String(), "ns/op") || stderr.Len() != 0 {
			t.Fatalf("unexpected filtered listing:\n%s\n%s", &output, &stderr)
		}
	}
}

type listingWriter struct{ err error }

func (w listingWriter) Write(p []byte) (int, error) { return 0, w.err }

func TestCatalogWriterFailures(t *testing.T) {
	want := errors.New("listing write failed")
	if err := printCatalog(listingWriter{want}, "all", capabilities{}); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
	if err := printCatalog(listingWriter{}, "all", capabilities{}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write = %v", err)
	}
}
