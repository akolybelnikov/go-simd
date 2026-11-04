package labapi

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

const CatalogSchema = "go-simd.lab-catalog/v1"

type CatalogResponse struct {
	Schema            string                `json:"schema"`
	Runtime           RuntimeContext        `json:"runtime"`
	FMA32Path         string                `json:"fma32_path,omitempty"`
	MeasurementNotice string                `json:"measurement_notice"`
	Operations        []OperationDescriptor `json:"operations"`
}

type ValueDescriptor struct {
	Name         string        `json:"name"`
	TypeID       string        `json:"type_id"`
	Shape        catalog.Shape `json:"shape"`
	ElementBytes int64         `json:"element_bytes"`
}

type ParameterDescriptor struct {
	Name     string `json:"name"`
	TypeID   string `json:"type_id"`
	Encoding string `json:"encoding"`
}

type StrategyDescriptor struct {
	ID           string               `json:"id"`
	Label        string               `json:"label"`
	UsesSIMD     bool                 `json:"uses_simd"`
	WorkerPolicy catalog.WorkerPolicy `json:"worker_policy"`
}

type OperationDescriptor struct {
	ID                 string                `json:"id"`
	Group              string                `json:"group"`
	Title              string                `json:"title"`
	Category           string                `json:"category"`
	TypeID             string                `json:"type_id"`
	ArithmeticID       string                `json:"arithmetic_id"`
	Expression         string                `json:"expression"`
	NumericalPolicy    string                `json:"numerical_policy"`
	ReferenceStrategy  string                `json:"reference_strategy_id"`
	ExecutionMode      string                `json:"execution_mode"`
	Inputs             []ValueDescriptor     `json:"inputs"`
	Output             ValueDescriptor       `json:"output"`
	Parameters         []ParameterDescriptor `json:"parameters,omitempty"`
	Strategies         []StrategyDescriptor  `json:"strategies"`
	SupportedSizes     []int                 `json:"supported_sizes"`
	LogicalBytesAtOne  int64                 `json:"logical_bytes_at_n_1"`
	LogicalBytesNotice string                `json:"logical_bytes_notice"`
}

// DescribeCatalog maps executable-free catalogue metadata into the stable lab
// contract. An empty group is treated as "all" for CLI convenience.
func DescribeCatalog(group string, runtime RuntimeContext, fma32Path string) (CatalogResponse, error) {
	if runtime.Float32Lanes < 1 || runtime.VectorBits != runtime.Float32Lanes*32 || runtime.GOMAXPROCS < 1 || runtime.GoVersion == "" || runtime.OS == "" || runtime.Arch == "" {
		return CatalogResponse{}, planError("invalid_runtime", "runtime", "valid measuring runtime identity is required")
	}
	if group == "" {
		group = "all"
	}
	response := CatalogResponse{
		Schema: CatalogSchema, Runtime: runtime, FMA32Path: fma32Path,
		MeasurementNotice: "Capabilities and future measurements describe this process and host, not a browser or remote client device.",
	}
	for _, op := range catalog.All() {
		if group != "all" && group != op.Group {
			continue
		}
		descriptor, err := describeOperation(op, runtime.Float32Lanes)
		if err != nil {
			return CatalogResponse{}, err
		}
		response.Operations = append(response.Operations, descriptor)
	}
	if len(response.Operations) == 0 {
		return CatalogResponse{}, planError("unknown_group", "group", "unknown operation group %q", group)
	}
	return response, nil
}

func describeOperation(op catalog.Operation, lanes int) (OperationDescriptor, error) {
	logicalBytes, err := op.LogicalBytes(1)
	if err != nil {
		return OperationDescriptor{}, fmt.Errorf("logical bytes for %s: %w", op.Group, err)
	}
	descriptor := OperationDescriptor{
		ID: op.ID, Group: op.Group, Title: op.Title, Category: op.Category,
		TypeID: op.TypeID, ArithmeticID: op.ArithmeticID, Expression: op.Expression,
		NumericalPolicy: op.NumericalPolicy, ReferenceStrategy: op.ReferenceVariantID,
		ExecutionMode: op.ExecutionMode, Output: describeValue(op.Output),
		SupportedSizes: benchutil.OperationSizes(op, lanes), LogicalBytesAtOne: logicalBytes,
		LogicalBytesNotice: "Effective logical work accounting, not measured DRAM traffic.",
	}
	for _, value := range op.Inputs {
		descriptor.Inputs = append(descriptor.Inputs, describeValue(value))
	}
	for _, parameter := range op.Parameters {
		descriptor.Parameters = append(descriptor.Parameters, ParameterDescriptor{Name: parameter.Name, TypeID: parameter.TypeID, Encoding: parameter.Encoding})
	}
	for _, strategy := range op.Variants {
		descriptor.Strategies = append(descriptor.Strategies, StrategyDescriptor{ID: strategy.ID, Label: strategy.Label, UsesSIMD: strategy.UsesSIMD, WorkerPolicy: strategy.Workers})
	}
	return descriptor, nil
}

func describeValue(value catalog.Value) ValueDescriptor {
	return ValueDescriptor{Name: value.Name, TypeID: value.TypeID, Shape: value.Shape, ElementBytes: value.ElementBytes}
}
