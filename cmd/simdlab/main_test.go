package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestPrivateWorkerRejectsInvalidRequest(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"_worker"}, strings.NewReader("{}"), &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "invalid worker request") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestHelpListsDeterministicCommands(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"help"}, strings.NewReader(""), &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "catalog") || !strings.Contains(stdout.String(), "inspect") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestCatalogAndPlanJSONCommands(t *testing.T) {
	var catalogOut, stderr bytes.Buffer
	if code := run([]string{"catalog", "-group", "addition"}, strings.NewReader(""), &catalogOut, &stderr); code != 0 {
		t.Fatalf("catalog code=%d stderr=%q", code, stderr.String())
	}
	var catalogue labapi.CatalogResponse
	if err := json.Unmarshal(catalogOut.Bytes(), &catalogue); err != nil || len(catalogue.Operations) != 1 {
		t.Fatalf("catalogue=%+v err=%v", catalogue, err)
	}
	request := labapi.PlanRequest{Group: "addition", Variants: []string{"scalar"}, Sizes: []int{catalogue.Operations[0].SupportedSizes[0]}, Count: 3, Benchtime: "1x"}
	requestData, _ := json.Marshal(request)
	var planOut bytes.Buffer
	stderr.Reset()
	if code := run([]string{"plan"}, bytes.NewReader(requestData), &planOut, &stderr); code != 0 {
		t.Fatalf("plan code=%d stderr=%q", code, stderr.String())
	}
	var plan labapi.ExperimentPlan
	if err := json.Unmarshal(planOut.Bytes(), &plan); err != nil || plan.ID == "" || plan.CaseCount != 1 {
		t.Fatalf("plan=%+v err=%v", plan, err)
	}
}

func TestPlanRejectsUnknownJSONFields(t *testing.T) {
	var stdout, stderr bytes.Buffer
	request := `{"group":"addition","variants":["scalar"],"sizes":[8],"count":3,"benchtime":"1x","command":"shell"}`
	if code := run([]string{"plan"}, strings.NewReader(request), &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "unknown field") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

func TestInspectCommandValidatesAndAnalyzesCapture(t *testing.T) {
	duration := int64(10)
	key := result.MeasurementKey{RunID: "run", EnvironmentID: "env", BackendID: "backend", HarnessID: "harness", Case: result.Case{OperationID: catalog.Add, TypeID: catalog.Float32, StrategyID: catalog.Scalar, Elements: 8, Input: result.InputIdentity{GeneratorID: "generated-float32-v1"}}}
	capture := evidence.Capture{
		Schema: evidence.Schema, Status: "complete", Group: "addition",
		Config:  evidence.Config{Count: 1, Benchtime: "1x", Bench: "."},
		Machine: evidence.Machine{GoVersion: "go-test", OS: "windows", Arch: "amd64", CPU: "test", Bits: 256, Lanes: 8, Procs: 4},
		Run:     result.Run{ID: "run", Environment: result.Environment{ID: "env"}, BackendID: "backend", HarnessID: "harness", SourceFormat: evidence.Schema, Observations: []result.Observation{{Key: key, Label: "Scalar", Repetition: 1, Iterations: 1, DurationNS: &duration, NanosecondsPerOp: 10, SourceBenchmark: "benchmark"}}},
	}
	path := filepath.Join(t.TempDir(), "capture.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := evidence.WriteCapture(file, capture); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"inspect", path}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	var inspection struct {
		Schema string `json:"schema"`
		Series []any  `json:"series"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &inspection); err != nil || inspection.Schema == "" || len(inspection.Series) != 1 {
		t.Fatalf("inspection=%+v err=%v output=%s", inspection, err, stdout.String())
	}
}
