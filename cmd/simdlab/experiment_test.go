package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/labworker"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == "_worker" {
		os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

func TestExperimentCLI(t *testing.T) {
	valid := labapi.PlanRequest{Group: "addition", Variants: []string{"scalar", "simd-partial"}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}
	plan, err := labapi.ResolveExperimentPlan(valid, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	require.NoError(t, err)
	for _, tt := range []struct {
		name, tool, arguments, confirmation string
		wantCode                            int
	}{
		{"confirmed", "plan_experiment", `{"group":"addition","variants":["scalar","simd-partial"],"sizes":[8],"count":1,"benchtime":"1x"}`, "run " + plan.ID + "\n", 0},
		{"exact flags", "", "", "run " + plan.ID + "\n", 0},
		{"wrong plan ID", "plan_experiment", `{"group":"addition","variants":["scalar"],"sizes":[8],"count":1,"benchtime":"1x"}`, "run " + plan.ID + "\n", 2},
		{"model cannot confirm", "plan_experiment", `{"group":"addition","variants":["scalar"],"sizes":[8],"count":1,"benchtime":"1x"}`, "", 2},
		{"execution tool unavailable", "run_experiment", `{}`, "run " + plan.ID + "\n", 1},
		{"worker option rejected", "plan_experiment", `{"group":"addition","variants":["scalar"],"sizes":[8],"count":1,"benchtime":"1x","path":"elsewhere"}`, "run " + plan.ID + "\n", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/api/chat" {
					http.NotFound(w, r)
					return
				}
				var req struct {
					Tools []struct{ Function struct{ Name string } }
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Tools) != 1 || req.Tools[0].Function.Name != "plan_experiment" {
					http.Error(w, "unexpected tool list", http.StatusBadRequest)
					return
				}
				_, _ = io.WriteString(w, `{"message":{"content":"I confirm this experiment.","tool_calls":[{"function":{"name":"`+tt.tool+`","arguments":`+tt.arguments+`}}]},"done":true}`)
			}))
			defer server.Close()
			root := t.TempDir()
			var stdout, stderr bytes.Buffer
			args := []string{"experiment", "-preview", "-artifacts", root}
			if tt.tool == "" {
				args = append(args, "-endpoint", "http://127.0.0.1:1", "-group", "addition", "-variants", "scalar,simd-partial", "-sizes", "8", "-count", "1", "-benchtime", "1x")
			} else {
				args = append(args, "-endpoint", server.URL, "Compare scalar and SIMD addition")
			}
			code := run(args, strings.NewReader(tt.confirmation), &stdout, &stderr)
			require.Equal(t, tt.wantCode, code, stderr.String())
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			if code != 0 {
				require.Empty(t, entries)
				require.Empty(t, stdout.String())
				return
			}
			require.Len(t, entries, 1)
			require.Contains(t, stderr.String(), "Type 'run "+plan.ID+"'")
			require.Contains(t, stderr.String(), "observations saved")
			var result struct {
				labworker.Manifest
				Evidence string
			}
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
			require.Equal(t, plan.ID, result.PlanID)
			require.Equal(t, "complete", result.Status)
			require.Equal(t, 2, result.Observations)
			require.FileExists(t, result.Evidence)
			stdout.Reset()
			require.Zero(t, run([]string{"inspect", result.Evidence}, strings.NewReader(""), &stdout, &stderr), stderr.String())

			interpretServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil || !bytes.Contains(body, []byte("series-1")) || bytes.Contains(body, []byte("measurements.json")) {
					http.Error(w, "expected evidence facts without file paths", http.StatusBadRequest)
					return
				}
				_, _ = io.WriteString(w, `{"message":{"tool_calls":[{"function":{"name":"select_evidence_facts","arguments":{"decision":"answer","fact_ids":["series-1"]}}}]},"done":true}`)
			}))
			defer interpretServer.Close()
			stdout.Reset()
			require.Zero(t, run([]string{"interpret", "-preview", "-endpoint", interpretServer.URL, result.Evidence}, strings.NewReader(""), &stdout, &stderr), stderr.String())
			require.Contains(t, stdout.String(), `"decision": "answer"`)
			require.Contains(t, stdout.String(), `"id": "series-1"`)
		})
	}
}
