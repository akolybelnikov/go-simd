package labservice

import (
	"io"

	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

// Deterministic contains model-free Explore, Plan and Inspect use cases. It
// owns no CLI text formatting and performs no measurement during these calls.
type Deterministic struct {
	machine platform.Info
	policy  labapi.PlanPolicy
}

func NewDeterministic(machine platform.Info, policy labapi.PlanPolicy) (*Deterministic, error) {
	if _, err := labapi.DescribeCatalog("all", labrun.Runtime(machine), machine.FMA32Path); err != nil {
		return nil, err
	}
	return &Deterministic{machine: machine, policy: policy}, nil
}

func (s *Deterministic) Catalog(group string) (labapi.CatalogResponse, error) {
	return labapi.DescribeCatalog(group, labrun.Runtime(s.machine), s.machine.FMA32Path)
}

func (s *Deterministic) Plan(request labapi.PlanRequest) (labapi.ExperimentPlan, error) {
	return labapi.ResolveExperimentPlan(request, labrun.Runtime(s.machine), s.policy)
}

func (s *Deterministic) Inspect(reader io.Reader, limits evidence.ReadLimits) (analysis.Interpretation, error) {
	capture, err := evidence.ReadCapture(reader, limits)
	if err != nil {
		return analysis.Interpretation{}, err
	}
	return analysis.InterpretCapture(capture)
}
