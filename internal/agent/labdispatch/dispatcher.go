package labdispatch

import (
	"context"
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/labapi"
)

// CatalogExplorer describes the deterministic catalog capability required by the dispatcher.
type CatalogExplorer interface {
	Catalog(group string) (labapi.CatalogResponse, error)
}

// Dispatcher describes the lab-specific implementation that converts validated agent tool calls into deterministic lab results.
type Dispatcher struct {
	catalogExplorer CatalogExplorer
}

func (d *Dispatcher) Dispatch(ctx context.Context, toolcall agent.ToolCall) (agent.Message, error) {
	if err := toolcall.Validate(); err != nil {
		return agent.Message{}, fmt.Errorf("invalid tool call: %w", err)
	}

	switch toolcall.Name {
	case "explore_catalog":
		return d.exploreCatalog(toolcall)
	default:
		return agent.Message{}, fmt.Errorf("unsupported tool call name %q", toolcall.Name)
	}
}

// NewDispatcher initializes an instance of the Dispatcher with the provided not-nil catalog explorer.
func NewDispatcher(catalogExplorer CatalogExplorer) (*Dispatcher, error) {
	if catalogExplorer == nil {
		return nil, fmt.Errorf("catalog explorer cannot be nil")
	}

	return &Dispatcher{catalogExplorer: catalogExplorer}, nil
}
