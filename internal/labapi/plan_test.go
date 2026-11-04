package labapi

import (
	"encoding/json"
	"errors"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func testRuntime() RuntimeContext {
	return RuntimeContext{GoVersion: "go1.27.3", OS: "linux", Arch: "amd64", VectorBits: 256, Float32Lanes: 8, GOMAXPROCS: 4}
}

func TestResolveExperimentPlanCanonicalAndImmutable(t *testing.T) {
	request := PlanRequest{Group: "addition", Variants: []string{catalog.SIMDPartial, catalog.Scalar}, Sizes: []int{9, 8}, Count: 2, Benchtime: "100ms"}
	plan, err := ResolveExperimentPlan(request, testRuntime(), DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Schema != PlanSchema || plan.ID == "" || plan.CaseCount != 4 || plan.MeasurementCount != 8 || len(plan.Cases) != 4 {
		t.Fatalf("unexpected plan header: %+v", plan)
	}
	wantNames := []string{"BenchmarkAddition/n=8/Scalar", "BenchmarkAddition/n=8/SIMDPartial", "BenchmarkAddition/n=9/Scalar", "BenchmarkAddition/n=9/SIMDPartial"}
	var names []string
	for _, c := range plan.Cases {
		names = append(names, c.Name)
		if c.OperationID != catalog.Add || c.TypeID != catalog.Float32 || c.LogicalBytes != int64(12*c.Elements) {
			t.Fatalf("unexpected case: %+v", c)
		}
	}
	if !slices.Equal(names, wantNames) {
		t.Fatalf("case order = %v, want %v", names, wantNames)
	}
	if err := VerifyExperimentPlan(plan); err != nil {
		t.Fatal(err)
	}

	reordered := request
	reordered.Variants = []string{catalog.Scalar, catalog.SIMDPartial}
	reordered.Sizes = []int{8, 9}
	second, err := ResolveExperimentPlan(reordered, testRuntime(), DefaultPlanPolicy())
	if err != nil || second.ID != plan.ID {
		t.Fatalf("equivalent request ID = %q, err = %v; want %q", second.ID, err, plan.ID)
	}

	plan.Count++
	if err := VerifyExperimentPlan(plan); errorCode(err) != "plan_modified" {
		t.Fatalf("modified plan error = %v", err)
	}
}

func TestResolveExperimentPlanRetainsExactRelevantInput(t *testing.T) {
	pattern := "shuffled"
	seed := uint64(0)
	threshold := uint32(0x80000000)
	irrelevantAlpha := math.Float32bits(99)
	request := PlanRequest{
		Group: "predicate", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 3, Benchtime: "10x",
		Input: &InputOptions{Pattern: &pattern, Seed: &seed, ThresholdBits: &threshold, AlphaBits: &irrelevantAlpha},
	}
	plan, err := ResolveExperimentPlan(request, testRuntime(), DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Input.GeneratorID != "predicate-float32-v1/shuffled/seed=0" || len(plan.Input.Parameters) != 1 || plan.Input.Parameters[0].Bits != threshold {
		t.Fatalf("unexpected input: %+v", plan.Input)
	}
	if plan.Input.Options.AlphaBits != math.Float32bits(1.5) {
		t.Fatal("irrelevant alpha affected normalized plan")
	}
	if len(plan.Warnings) != 1 || plan.Warnings[0].Code != "fixed_iterations" {
		t.Fatalf("warnings = %+v", plan.Warnings)
	}
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ExperimentPlan
	if err := json.Unmarshal(data, &decoded); err != nil || VerifyExperimentPlan(decoded) != nil {
		t.Fatalf("round trip failed: %v", err)
	}
}

func TestResolveExperimentPlanWarnings(t *testing.T) {
	plan, err := ResolveExperimentPlan(PlanRequest{
		Group: "addition", Variants: []string{catalog.SIMDParallelScalarTail}, Sizes: []int{8}, Count: 1, Benchtime: "50ms",
	}, testRuntime(), DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	var codes []string
	for _, warning := range plan.Warnings {
		codes = append(codes, warning.Code)
	}
	if !slices.Equal(codes, []string{"low_repetition_count", "parallel_vs_serial"}) {
		t.Fatalf("warning codes = %v", codes)
	}
}

func TestResolveExperimentPlanRejectsInvalidRequests(t *testing.T) {
	valid := PlanRequest{Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 3, Benchtime: "100ms"}
	tests := []struct {
		name   string
		mutate func(*PlanRequest, *RuntimeContext, *PlanPolicy)
		code   string
	}{
		{"unknown group", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Group = "missing" }, "unknown_group"},
		{"missing size", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Sizes = nil }, "required"},
		{"unsupported size", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Sizes = []int{3} }, "unsupported_size"},
		{"duplicate size", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Sizes = []int{8, 8} }, "duplicate_size"},
		{"missing variant", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Variants = nil }, "required"},
		{"unknown variant", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Variants = []string{"missing"} }, "unknown_variant"},
		{"duplicate variant", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) {
			r.Variants = []string{catalog.Scalar, catalog.Scalar}
		}, "duplicate_variant"},
		{"bad count", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Count = 0 }, "invalid_count"},
		{"count policy", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Count = 11 }, "policy_exceeded"},
		{"bad benchtime", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Benchtime = "0s" }, "invalid_benchtime"},
		{"duration policy", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Benchtime = "2s" }, "policy_exceeded"},
		{"iteration policy", func(r *PlanRequest, _ *RuntimeContext, _ *PlanPolicy) { r.Benchtime = "1000001x" }, "policy_exceeded"},
		{"runtime", func(_ *PlanRequest, rt *RuntimeContext, _ *PlanPolicy) { rt.VectorBits = 128 }, "invalid_runtime"},
		{"policy", func(_ *PlanRequest, _ *RuntimeContext, p *PlanPolicy) { p.MaxCases = 0 }, "invalid_policy"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, runtime, policy := valid, testRuntime(), DefaultPlanPolicy()
			test.mutate(&request, &runtime, &policy)
			_, err := ResolveExperimentPlan(request, runtime, policy)
			if errorCode(err) != test.code {
				t.Fatalf("error = %v, code = %q; want %q", err, errorCode(err), test.code)
			}
		})
	}
}

func TestResolveExperimentPlanRejectsInvalidOperationInputAndLimits(t *testing.T) {
	lower, upper := math.Float32bits(2), math.Float32bits(1)
	_, err := ResolveExperimentPlan(PlanRequest{
		Group: "clamp", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 3, Benchtime: "100ms",
		Input: &InputOptions{LowerBits: &lower, UpperBits: &upper},
	}, testRuntime(), DefaultPlanPolicy())
	if errorCode(err) != "invalid_input" {
		t.Fatalf("invalid clamp error = %v", err)
	}

	policy := DefaultPlanPolicy()
	policy.MaxCases = 1
	_, err = ResolveExperimentPlan(PlanRequest{Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8, 16}, Count: 3, Benchtime: "100ms"}, testRuntime(), policy)
	if errorCode(err) != "policy_exceeded" {
		t.Fatalf("case policy error = %v", err)
	}

	policy = DefaultPlanPolicy()
	policy.MaxMeasurements = 2
	_, err = ResolveExperimentPlan(PlanRequest{Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 3, Benchtime: "100ms"}, testRuntime(), policy)
	if errorCode(err) != "policy_exceeded" {
		t.Fatalf("measurement policy error = %v", err)
	}
}

func TestReferenceVariantPlanForEveryOperation(t *testing.T) {
	for _, op := range catalog.All() {
		t.Run(op.Group, func(t *testing.T) {
			plan, err := ResolveExperimentPlan(PlanRequest{Group: op.Group, Variants: []string{op.ReferenceVariantID}, Sizes: []int{8}, Count: 3, Benchtime: "100ms"}, testRuntime(), DefaultPlanPolicy())
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Cases) != 1 || plan.Cases[0].StrategyID != op.ReferenceVariantID || VerifyExperimentPlan(plan) != nil {
				t.Fatalf("unexpected plan: %+v", plan)
			}
		})
	}
}

func TestPlanPolicyDurationEncodingDoesNotAffectPlanID(t *testing.T) {
	request := PlanRequest{Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 3, Benchtime: "100ms"}
	policy := DefaultPlanPolicy()
	plan, err := ResolveExperimentPlan(request, testRuntime(), policy)
	if err != nil {
		t.Fatal(err)
	}
	policy.MaxBenchtime = 2 * time.Second
	second, err := ResolveExperimentPlan(request, testRuntime(), policy)
	if err != nil || second.ID != plan.ID {
		t.Fatalf("policy limit changed accepted plan identity: %q vs %q, err %v", plan.ID, second.ID, err)
	}
}

func TestRevalidateExperimentPlan(t *testing.T) {
	plan, err := ResolveExperimentPlan(PlanRequest{Group: "addition", Variants: []string{"scalar"}, Sizes: []int{16}, Count: 3, Benchtime: "1x"}, testRuntime(), DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if err := RevalidateExperimentPlan(plan, testRuntime(), DefaultPlanPolicy()); err != nil {
		t.Fatalf("revalidate: %v", err)
	}

	otherRuntime := testRuntime()
	otherRuntime.GOMAXPROCS++
	if err := RevalidateExperimentPlan(plan, otherRuntime, DefaultPlanPolicy()); errorCode(err) != "runtime_mismatch" {
		t.Fatalf("runtime mismatch error = %v", err)
	}

	policy := DefaultPlanPolicy()
	policy.ID = "other/v1"
	if err := RevalidateExperimentPlan(plan, testRuntime(), policy); errorCode(err) != "policy_mismatch" {
		t.Fatalf("policy mismatch error = %v", err)
	}
}

func errorCode(err error) string {
	var target *PlanError
	if errors.As(err, &target) {
		return target.Code
	}
	return ""
}
