package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// These fixtures characterize the legacy parser and table before M1 extracts
// their responsibilities. They are synthetic examples, not timing assertions.
func TestLegacyReportFixtures(t *testing.T) {
	for _, name := range []string{"addition", "computation", "filtered"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", name+".raw"))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("testdata", name+".want.txt"))
			if err != nil {
				t.Fatal(err)
			}
			got, err := comparison(string(raw), name)
			if err != nil {
				t.Fatal(err)
			}
			if normalizeTable(got) != normalizeTable(string(want)) {
				t.Fatalf("report differs from compatibility fixture\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

// Alignment padding is incidental; blank lines, row/column order, precision,
// signs, and missing-value markers are part of the compatibility contract.
func normalizeTable(value string) string {
	lines := strings.Split(strings.TrimSpace(value), "\n")
	for i, line := range lines {
		lines[i] = strings.Join(strings.Fields(line), " ")
	}
	return strings.Join(lines, "\n")
}

func TestReportRejectsInvalidOptionsBeforeArtifacts(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{"unknown group", []string{"-group=unknown"}, `unknown group "unknown"`},
		{"zero count", []string{"-count=0"}, "count must be positive"},
		{"negative count", []string{"-count=-1"}, "count must be positive"},
		{"positional argument", []string{"extra"}, "unexpected positional arguments"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "reports")
			args := append([]string{"-output", output}, tt.args...)
			var stdout bytes.Buffer
			err := run(args, &stdout)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("run() error = %v, want %q", err, tt.want)
			}
			if stdout.Len() != 0 {
				t.Fatalf("unexpected benchmark progress: %s", stdout.String())
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("invalid options created artifacts: %v", err)
			}
		})
	}
}

func TestReportHelpCompatibility(t *testing.T) {
	// A separate process captures flag's real stderr and main's exit behavior
	// without changing process-wide stderr or refactoring production code in M0.
	command := exec.Command(os.Args[0], "-test.run=^TestReportCommandProcess$", "--", "-h")
	command.Env = append(os.Environ(), "GO_SIMD_M0_REPORT_HELP=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("help failed: %v\n%s", err, stderr.String())
	}
	want, err := os.ReadFile(filepath.Join("testdata", "help.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 || strings.ReplaceAll(stderr.String(), "\r\n", "\n") != string(want) {
		t.Fatalf("help differs from fixture\nstdout: %q\nstderr: %q\nwant: %q", stdout.String(), stderr.String(), want)
	}
}

func TestReportCommandProcess(t *testing.T) {
	if os.Getenv("GO_SIMD_M0_REPORT_HELP") != "1" {
		return
	}
	separator := slices.Index(os.Args, "--")
	if separator < 0 {
		t.Fatal("missing command argument separator")
	}
	os.Args = append([]string{"simdreport"}, os.Args[separator+1:]...)
	main()
	os.Exit(0)
}
