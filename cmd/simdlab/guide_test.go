package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/labworker"
	"github.com/stretchr/testify/require"
)

func TestStarterAndGuideSharePlanAndEvidence(t *testing.T) {
	var stdout, stderr bytes.Buffer
	direct := t.TempDir()
	require.Zero(t, run([]string{"run", "-artifacts", direct, "addition"}, strings.NewReader(""), &stdout, &stderr), stderr.String())
	directPlan, _ := checkStarterFiles(t, direct, stdout.String())

	stdout.Reset()
	stderr.Reset()
	guided := t.TempDir()
	input := strings.Join([]string{guideExample, "yes", "Which method took less time? Were the repeats consistent?", "quit"}, "\n")
	require.Zero(t, run([]string{"ask", "-endpoint", "http://127.0.0.1:1", "-artifacts", guided}, strings.NewReader(input), &stdout, &stderr), stderr.String())
	guidedPlan, evidence := checkStarterFiles(t, guided, stdout.String())
	require.Equal(t, directPlan, guidedPlan)
	require.Contains(t, stdout.String(), "Run this comparison? [yes/no]")
	require.Equal(t, 2, strings.Count(stdout.String(), "Median per batch"))
	require.Contains(t, stdout.String(), "3 repeats, ")
	require.NotContains(t, stdout.String(), "sha256:")

	stdout.Reset()
	require.Zero(t, run([]string{"ask", "-evidence", evidence}, strings.NewReader("Which method was faster?\nquit\n"), &stdout, &stderr))
	require.Contains(t, stdout.String(), "Median per batch")
	require.Contains(t, stdout.String(), evidence)
}

func checkStarterFiles(t *testing.T, root, output string) (string, string) {
	t.Helper()
	dirs, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Len(t, dirs, 1)
	job := filepath.Join(root, dirs[0].Name())
	data, err := os.ReadFile(filepath.Join(job, "manifest.json"))
	require.NoError(t, err)
	var manifest labworker.Manifest
	require.NoError(t, json.Unmarshal(data, &manifest))
	require.Equal(t, "complete", manifest.Status)
	require.Equal(t, 6, manifest.Observations)
	require.NotEmpty(t, manifest.PlanID)
	reportPath := filepath.Join(job, "report.md")
	report, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	require.Contains(t, string(report), "[measurements.json](measurements.json)")
	require.Contains(t, string(report), "3 observations, range")
	require.Contains(t, output, "Median per batch")
	require.Contains(t, output, "This run used generated data on this machine.")
	require.Contains(t, output, reportPath)
	evidence := filepath.Join(job, "measurements.json")
	require.FileExists(t, evidence)
	require.Contains(t, output, evidence)
	return manifest.PlanID, evidence
}

func TestGuideDiscoverClarifyAndDecline(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := t.TempDir()
	input := strings.Join([]string{
		"Can I compare ways of summing an array?",
		"Can you benchmark addition from my CSV?",
		"Compare ordinary and vectorized addition for 1,003 values.",
		"1,024",
		"no",
		"Which took less time, and by how much?",
		"quit",
	}, "\n")
	require.Zero(t, run([]string{"ask", "-endpoint", "http://127.0.0.1:1", "-artifacts", root}, strings.NewReader(input), &stdout, &stderr), stderr.String())
	require.Contains(t, stdout.String(), "Vectorized sum (four accumulators)")
	require.Contains(t, stdout.String(), "cannot benchmark your files")
	require.Contains(t, stdout.String(), "Which supported size should I use?")
	require.Contains(t, stdout.String(), "1024 values per input array")
	require.Contains(t, stdout.String(), "Comparison canceled.")
	require.Contains(t, stdout.String(), "No saved measurements yet.")
	dirs, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, dirs)
}

func TestGuideModelNeedsTerminalApprovalAndFallsBackOnTimeout(t *testing.T) {
	for _, tc := range []struct {
		name, modelReply, timeout, want string
	}{
		{"unconfirmed", `{"message":{"content":"yes","tool_calls":[{"function":{"name":"route_lab_request","arguments":{"action":"compare","group":"addition"}}}]},"done":true}`, "2s", "Run this comparison? [yes/no]"},
		{"invalid method", `{"message":{"tool_calls":[{"function":{"name":"route_lab_request","arguments":{"action":"compare","group":"addition","methods":["scalar","invented"]}}}]},"done":true}`, "2s", "Which two supported methods"},
		{"timeout", "", "100ms", "simdlab run addition"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				if tc.modelReply == "" {
					<-r.Context().Done()
					return
				}
				_, _ = io.WriteString(w, tc.modelReply)
			}))
			defer server.Close()
			var stdout, stderr bytes.Buffer
			root := t.TempDir()
			input := strings.NewReader("Could you time two variants of adding floats?\n")
			require.Zero(t, run([]string{"ask", "-endpoint", server.URL, "-timeout", tc.timeout, "-artifacts", root}, input, &stdout, &stderr), stderr.String())
			require.Contains(t, stdout.String(), tc.want)
			dirs, err := os.ReadDir(root)
			require.NoError(t, err)
			require.Empty(t, dirs)
		})
	}
}
