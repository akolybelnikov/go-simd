// Package labapi defines versioned, executable-data-free contracts for guided
// laboratory workflows. It does not expose kernel callbacks or filesystem paths.
package labapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
)

const PlanSchema = "go-simd.lab-plan/v1"

const defaultPolicyID = "guided-local-preview/v1"

// PlanError is a stable machine-readable validation failure. Message remains
// human-readable, while Code and Field are suitable for a CLI or agent tool.
type PlanError struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func (e *PlanError) Error() string { return e.Message }

func planError(code, field, format string, args ...any) error {
	return &PlanError{Code: code, Field: field, Message: fmt.Sprintf(format, args...)}
}

// InputOptions uses pointers so an omitted field is distinguishable from an
// explicitly requested zero bit-pattern or seed. ResolveExperimentPlan merges
// omitted values with the laboratory defaults and removes irrelevant settings.
type InputOptions struct {
	FMAInput      *string `json:"fma_input,omitempty"`
	DotInput      *string `json:"dot_input,omitempty"`
	SumInput      *string `json:"sum_input,omitempty"`
	Pattern       *string `json:"pattern,omitempty"`
	Seed          *uint64 `json:"seed,omitempty"`
	AlphaBits     *uint32 `json:"alpha_bits,omitempty"`
	BetaBits      *uint32 `json:"beta_bits,omitempty"`
	ThresholdBits *uint32 `json:"threshold_bits,omitempty"`
	LowerBits     *uint32 `json:"lower_bits,omitempty"`
	UpperBits     *uint32 `json:"upper_bits,omitempty"`
}

// PlanRequest contains only explicit operation selections. Sizes and variants
// are required so a guided request cannot silently expand into a full sweep.
type PlanRequest struct {
	Group     string        `json:"group"`
	Variants  []string      `json:"variants"`
	Sizes     []int         `json:"sizes"`
	Count     int           `json:"count"`
	Benchtime string        `json:"benchtime"`
	Input     *InputOptions `json:"input,omitempty"`
}

// RuntimeContext identifies the machine/process whose supported size boundaries
// and later measurements the plan describes. It never describes a browser client.
type RuntimeContext struct {
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	VectorBits   int    `json:"vector_bits"`
	Float32Lanes int    `json:"float32_lanes"`
	GOMAXPROCS   int    `json:"gomaxprocs"`
	Emulated     bool   `json:"emulated"`
}

// PlanPolicy applies bounded planning limits before any artifacts or kernel
// fixtures are created. Execution must reapply its current policy to the plan.
type PlanPolicy struct {
	ID                 string        `json:"id"`
	MaxCases           int           `json:"max_cases"`
	MaxCount           int           `json:"max_count"`
	MaxMeasurements    int           `json:"max_measurements"`
	MaxBenchtime       time.Duration `json:"max_benchtime"`
	MaxFixedIterations uint64        `json:"max_fixed_iterations"`
}

func DefaultPlanPolicy() PlanPolicy {
	return PlanPolicy{
		ID: defaultPolicyID, MaxCases: 32, MaxCount: 10,
		MaxMeasurements: 200, MaxBenchtime: time.Second,
		MaxFixedIterations: 1_000_000,
	}
}

func (p PlanPolicy) validate() error {
	if p.ID == "" || p.MaxCases < 1 || p.MaxCount < 1 || p.MaxMeasurements < 1 || p.MaxBenchtime <= 0 || p.MaxFixedIterations < 1 {
		return planError("invalid_policy", "policy", "plan policy must have an ID and positive limits")
	}
	return nil
}

// ResolvedInput is the exact normalized input configuration retained by the
// plan. Irrelevant request fields are reset to defaults before hashing.
type ResolvedInput struct {
	GeneratorID string               `json:"generator_id"`
	Parameters  []ParameterBits      `json:"parameters,omitempty"`
	Options     ResolvedInputOptions `json:"options"`
}

type ParameterBits struct {
	Name     string `json:"name"`
	TypeID   string `json:"type_id"`
	Encoding string `json:"encoding"`
	Bits     uint32 `json:"bits"`
}

// ResolvedInputOptions is complete rather than sparse so it can later be mapped
// to input.Options without consulting mutable defaults.
type ResolvedInputOptions struct {
	FMAInput      string `json:"fma_input"`
	DotInput      string `json:"dot_input"`
	SumInput      string `json:"sum_input"`
	Pattern       string `json:"pattern"`
	Seed          uint64 `json:"seed"`
	AlphaBits     uint32 `json:"alpha_bits"`
	BetaBits      uint32 `json:"beta_bits"`
	ThresholdBits uint32 `json:"threshold_bits"`
	LowerBits     uint32 `json:"lower_bits"`
	UpperBits     uint32 `json:"upper_bits"`
}

type PlannedCase struct {
	OutputElements int                  `json:"output_elements,omitempty"`
	Name           string               `json:"name"`
	OperationID    string               `json:"operation_id"`
	TypeID         string               `json:"type_id"`
	ArithmeticID   string               `json:"arithmetic_id"`
	StrategyID     string               `json:"strategy_id"`
	StrategyLabel  string               `json:"strategy_label"`
	Elements       int                  `json:"elements"`
	LogicalBytes   int64                `json:"logical_bytes"`
	UsesSIMD       bool                 `json:"uses_simd"`
	WorkerPolicy   catalog.WorkerPolicy `json:"worker_policy"`
}

type PlanWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ExperimentPlan is immutable by identity: ID is the SHA-256 of every other
// serialized field. VerifyExperimentPlan detects mutation before execution.
type ExperimentPlan struct {
	Schema            string         `json:"schema"`
	ID                string         `json:"id"`
	PolicyID          string         `json:"policy_id"`
	Group             string         `json:"group"`
	OperationID       string         `json:"operation_id"`
	TypeID            string         `json:"type_id"`
	ArithmeticID      string         `json:"arithmetic_id"`
	NumericalPolicy   string         `json:"numerical_policy"`
	Runtime           RuntimeContext `json:"runtime"`
	Input             ResolvedInput  `json:"input"`
	Count             int            `json:"count"`
	Benchtime         string         `json:"benchtime"`
	Cases             []PlannedCase  `json:"cases"`
	CaseCount         int            `json:"case_count"`
	MeasurementCount  int            `json:"measurement_count"`
	MeasurementNotice string         `json:"measurement_notice"`
	Warnings          []PlanWarning  `json:"warnings,omitempty"`
}

// ResolveExperimentPlan validates and normalizes a bounded request without
// allocating fixtures, executing kernels, or creating artifacts.
func ResolveExperimentPlan(request PlanRequest, runtime RuntimeContext, policy PlanPolicy) (ExperimentPlan, error) {
	if err := policy.validate(); err != nil {
		return ExperimentPlan{}, err
	}
	if runtime.Float32Lanes < 1 || runtime.VectorBits < 1 || runtime.VectorBits != runtime.Float32Lanes*32 {
		return ExperimentPlan{}, planError("invalid_runtime", "runtime", "runtime vector bits must equal float32 lanes * 32")
	}
	if runtime.GOMAXPROCS < 1 || runtime.OS == "" || runtime.Arch == "" || runtime.GoVersion == "" {
		return ExperimentPlan{}, planError("invalid_runtime", "runtime", "runtime OS, architecture, Go version, and GOMAXPROCS are required")
	}
	op, ok := catalog.ByGroup(request.Group)
	if !ok {
		return ExperimentPlan{}, planError("unknown_group", "group", "unknown operation group %q", request.Group)
	}
	if request.Count < 1 {
		return ExperimentPlan{}, planError("invalid_count", "count", "count must be positive")
	}
	if request.Count > policy.MaxCount {
		return ExperimentPlan{}, planError("policy_exceeded", "count", "count %d exceeds policy maximum %d", request.Count, policy.MaxCount)
	}
	fixed, fixedIterations, duration, err := validateBenchtime(request.Benchtime)
	if err != nil {
		return ExperimentPlan{}, err
	}
	if fixed && fixedIterations > policy.MaxFixedIterations {
		return ExperimentPlan{}, planError("policy_exceeded", "benchtime", "fixed iterations %d exceed policy maximum %d", fixedIterations, policy.MaxFixedIterations)
	}
	if !fixed && duration > policy.MaxBenchtime {
		return ExperimentPlan{}, planError("policy_exceeded", "benchtime", "benchtime %s exceeds policy maximum %s", duration, policy.MaxBenchtime)
	}

	sizes, err := resolveSizes(op, runtime.Float32Lanes, request.Sizes)
	if err != nil {
		return ExperimentPlan{}, err
	}
	variants, err := resolveVariants(op, request.Variants)
	if err != nil {
		return ExperimentPlan{}, err
	}
	options := mergeInputOptions(request.Input)
	options = normalizeInputOptions(op, options)
	if err := options.ValidateFor(op); err != nil {
		return ExperimentPlan{}, planError("invalid_input", "input", "%v", err)
	}

	caseCount := len(sizes) * len(variants)
	if caseCount > policy.MaxCases {
		return ExperimentPlan{}, planError("policy_exceeded", "cases", "resolved case count %d exceeds policy maximum %d", caseCount, policy.MaxCases)
	}
	if caseCount > math.MaxInt/request.Count || caseCount*request.Count > policy.MaxMeasurements {
		return ExperimentPlan{}, planError("policy_exceeded", "measurements", "resolved measurement count exceeds policy maximum %d", policy.MaxMeasurements)
	}

	plan := ExperimentPlan{
		Schema: PlanSchema, PolicyID: policy.ID, Group: op.Group,
		OperationID: op.ID, TypeID: op.TypeID, ArithmeticID: op.ArithmeticID,
		NumericalPolicy: op.NumericalPolicy, Runtime: runtime,
		Input: resolvedInput(op, options), Count: request.Count,
		Benchtime: request.Benchtime, CaseCount: caseCount,
		MeasurementCount:  caseCount * request.Count,
		MeasurementNotice: "Measurements describe the machine and process identified by runtime, not a remote browser or client device.",
	}
	for _, size := range sizes {
		logicalBytes, err := op.LogicalBytes(size)
		if err != nil {
			return ExperimentPlan{}, planError("invalid_size", "sizes", "%v", err)
		}
		outputElements := 0
		if op.Output.Shape == catalog.Blocks8 {
			outputElements, _ = op.OutputElements(size)
		}
		for _, variant := range variants {
			plan.Cases = append(plan.Cases, PlannedCase{
				Name:        fmt.Sprintf("%s/n=%d/%s", op.BenchmarkName, size, variant.Label),
				OperationID: op.ID, TypeID: op.TypeID, ArithmeticID: op.ArithmeticID,
				StrategyID: variant.ID, StrategyLabel: variant.Label, Elements: size,
				LogicalBytes: logicalBytes, UsesSIMD: variant.UsesSIMD,
				WorkerPolicy: variant.Workers, OutputElements: outputElements,
			})
		}
	}
	if fixed {
		plan.Warnings = append(plan.Warnings, PlanWarning{Code: "fixed_iterations", Message: "Fixed iteration mode is useful for plumbing checks but short batches may be unreliable performance evidence."})
	}
	if request.Count < 3 {
		plan.Warnings = append(plan.Warnings, PlanWarning{Code: "low_repetition_count", Message: "Fewer than three repetitions provide very limited variability evidence."})
	}
	if slices.ContainsFunc(variants, func(v catalog.Variant) bool { return v.Workers == catalog.PerCallGOMAXPROCS }) {
		plan.Warnings = append(plan.Warnings, PlanWarning{Code: "parallel_vs_serial", Message: "Parallel strategy comparisons include scheduling and concurrency effects in addition to SIMD."})
	}
	plan.ID, err = experimentPlanID(plan)
	if err != nil {
		return ExperimentPlan{}, err
	}
	return plan, nil
}

func validateBenchtime(value string) (fixed bool, iterations uint64, duration time.Duration, err error) {
	if before, ok := strings.CutSuffix(value, "x"); ok {
		n, parseErr := strconv.ParseUint(before, 10, 64)
		if parseErr != nil || n < 1 {
			return false, 0, 0, planError("invalid_benchtime", "benchtime", "invalid benchtime %q: use a positive duration or iteration count Nx", value)
		}
		return true, n, 0, nil
	}
	d, parseErr := time.ParseDuration(value)
	if parseErr != nil || d <= 0 {
		return false, 0, 0, planError("invalid_benchtime", "benchtime", "invalid benchtime %q: use a positive duration or iteration count Nx", value)
	}
	return false, 0, d, nil
}

func resolveSizes(op catalog.Operation, width int, requested []int) ([]int, error) {
	if len(requested) == 0 {
		return nil, planError("required", "sizes", "at least one explicit supported size is required")
	}
	supported := benchutil.OperationSizes(op, width)
	seen := make(map[int]bool, len(requested))
	resolved := slices.Clone(requested)
	for _, size := range resolved {
		if seen[size] {
			return nil, planError("duplicate_size", "sizes", "size %d is repeated", size)
		}
		seen[size] = true
		if !slices.Contains(supported, size) {
			return nil, planError("unsupported_size", "sizes", "size %d is not supported for group %q", size, op.Group)
		}
	}
	slices.Sort(resolved)
	return resolved, nil
}

func resolveVariants(op catalog.Operation, requested []string) ([]catalog.Variant, error) {
	if len(requested) == 0 {
		return nil, planError("required", "variants", "at least one explicit strategy ID is required")
	}
	seen := make(map[string]bool, len(requested))
	wanted := make(map[string]bool, len(requested))
	for _, id := range requested {
		if seen[id] {
			return nil, planError("duplicate_variant", "variants", "strategy %q is repeated", id)
		}
		seen[id], wanted[id] = true, true
	}
	var resolved []catalog.Variant
	for _, variant := range op.Variants {
		if wanted[variant.ID] {
			resolved = append(resolved, variant)
			delete(wanted, variant.ID)
		}
	}
	if len(wanted) != 0 {
		var unknown []string
		for id := range wanted {
			unknown = append(unknown, id)
		}
		slices.Sort(unknown)
		return nil, planError("unknown_variant", "variants", "unknown strategy ID(s) for group %q: %s", op.Group, strings.Join(unknown, ", "))
	}
	return resolved, nil
}

func mergeInputOptions(request *InputOptions) input.Options {
	resolved := input.DefaultOptions()
	if request == nil {
		return resolved
	}
	if request.FMAInput != nil {
		resolved.FMAInput = *request.FMAInput
	}
	if request.DotInput != nil {
		resolved.DotInput = *request.DotInput
	}
	if request.SumInput != nil {
		resolved.SumInput = *request.SumInput
	}
	if request.Pattern != nil {
		resolved.Pattern = *request.Pattern
	}
	if request.Seed != nil {
		resolved.Seed = *request.Seed
	}
	if request.AlphaBits != nil {
		resolved.Alpha = *request.AlphaBits
	}
	if request.BetaBits != nil {
		resolved.Beta = *request.BetaBits
	}
	if request.ThresholdBits != nil {
		resolved.Threshold = *request.ThresholdBits
	}
	if request.LowerBits != nil {
		resolved.Lower = *request.LowerBits
	}
	if request.UpperBits != nil {
		resolved.Upper = *request.UpperBits
	}
	return resolved
}

func normalizeInputOptions(op catalog.Operation, requested input.Options) input.Options {
	resolved := input.DefaultOptions()
	for _, parameter := range op.Parameters {
		switch parameter.Name {
		case "alpha":
			resolved.Alpha = requested.Alpha
		case "beta":
			resolved.Beta = requested.Beta
		case "threshold":
			resolved.Threshold = requested.Threshold
		case "lower":
			resolved.Lower = requested.Lower
		case "upper":
			resolved.Upper = requested.Upper
		}
	}
	if input.IsPredicate(op) || op.ID == catalog.Clamp {
		resolved.Pattern, resolved.Seed = requested.Pattern, requested.Seed
	}
	switch op.ID {
	case catalog.Sum, catalog.Sum8:
		resolved.SumInput = requested.SumFixture()
	case catalog.Dot, catalog.DotFMA:
		resolved.DotInput = requested.DotFixture()
	case catalog.AffineFMA, catalog.MulAdd:
		resolved.FMAInput = requested.FMAFixture()
	}
	return resolved
}

func resolvedInput(op catalog.Operation, options input.Options) ResolvedInput {
	parameters := make([]ParameterBits, 0, len(op.Parameters))
	for _, parameter := range op.Parameters {
		parameters = append(parameters, ParameterBits{
			Name: parameter.Name, TypeID: parameter.TypeID,
			Encoding: parameter.Encoding, Bits: options.Bits(parameter.Name),
		})
	}
	return ResolvedInput{
		GeneratorID: opGenerator(op, options), Parameters: parameters,
		Options: ResolvedInputOptions{
			FMAInput: options.FMAFixture(), DotInput: options.DotFixture(), SumInput: options.SumFixture(),
			Pattern: options.Pattern, Seed: options.Seed,
			AlphaBits: options.Alpha, BetaBits: options.Beta, ThresholdBits: options.Threshold,
			LowerBits: options.Lower, UpperBits: options.Upper,
		},
	}
}

func opGenerator(op catalog.Operation, options input.Options) string {
	return options.Generator(op)
}

// VerifyExperimentPlan rejects unsupported schemas, missing IDs, and any change
// to a confirmed plan's normalized content.
func VerifyExperimentPlan(plan ExperimentPlan) error {
	if plan.Schema != PlanSchema {
		return planError("unsupported_schema", "schema", "unsupported plan schema %q", plan.Schema)
	}
	if plan.ID == "" {
		return planError("invalid_plan_id", "id", "plan ID is required")
	}
	want, err := experimentPlanID(plan)
	if err != nil {
		return err
	}
	if plan.ID != want {
		return planError("plan_modified", "id", "plan content does not match confirmed plan ID")
	}
	return nil
}

// RevalidateExperimentPlan reapplies the current execution policy and runtime
// to a confirmed plan. This prevents a valid plan created under broader limits
// or for a different measuring process from being replayed at execution time.
func RevalidateExperimentPlan(plan ExperimentPlan, runtime RuntimeContext, policy PlanPolicy) error {
	if err := VerifyExperimentPlan(plan); err != nil {
		return err
	}
	if plan.PolicyID != policy.ID {
		return planError("policy_mismatch", "policy_id", "plan policy %q does not match execution policy %q", plan.PolicyID, policy.ID)
	}
	if plan.Runtime != runtime {
		return planError("runtime_mismatch", "runtime", "plan runtime does not match the measuring process")
	}
	variants := make([]string, 0, len(plan.Cases))
	sizes := make([]int, 0, len(plan.Cases))
	for _, c := range plan.Cases {
		if !slices.Contains(variants, c.StrategyID) {
			variants = append(variants, c.StrategyID)
		}
		if !slices.Contains(sizes, c.Elements) {
			sizes = append(sizes, c.Elements)
		}
	}
	o := plan.Input.Options
	request := PlanRequest{
		Group: plan.Group, Variants: variants, Sizes: sizes,
		Count: plan.Count, Benchtime: plan.Benchtime,
		Input: &InputOptions{
			FMAInput: &o.FMAInput, DotInput: &o.DotInput, SumInput: &o.SumInput,
			Pattern: &o.Pattern, Seed: &o.Seed,
			AlphaBits: &o.AlphaBits, BetaBits: &o.BetaBits,
			ThresholdBits: &o.ThresholdBits, LowerBits: &o.LowerBits, UpperBits: &o.UpperBits,
		},
	}
	resolved, err := ResolveExperimentPlan(request, runtime, policy)
	if err != nil {
		return err
	}
	if resolved.ID != plan.ID {
		return planError("plan_inconsistent", "plan", "plan does not match the canonical plan for its selections")
	}
	return nil
}

func experimentPlanID(plan ExperimentPlan) (string, error) {
	plan.ID = ""
	data, err := json.Marshal(plan)
	if err != nil {
		return "", planError("encoding_failed", "plan", "encoding normalized plan: %v", err)
	}
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}
