package labdispatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/akolybelnikov/go-simd/internal/agent"
)

type catalogArguments struct {
	Group string `json:"group"`
}

func decodeCatalogArguments(arguments []byte) (catalogArguments, error) {
	decoder := json.NewDecoder(bytes.NewReader(arguments))
	decoder.DisallowUnknownFields()

	val := catalogArguments{}
	if err := decoder.Decode(&val); err != nil {
		return catalogArguments{}, fmt.Errorf("failed to decode the arguments: %w", err)
	}

	if val.Group == "" {
		return catalogArguments{}, fmt.Errorf("group cannot be missing in the arguments")
	}

	trail := json.RawMessage{}
	if err := decoder.Decode(&trail); err == nil {
		return catalogArguments{}, fmt.Errorf("invalid trailing data")
	} else {
		if !errors.Is(err, io.EOF) {
			return catalogArguments{}, fmt.Errorf("error decoding trailing data: %w", err)
		}
	}

	return val, nil
}

func (d *Dispatcher) exploreCatalog(tc agent.ToolCall) (agent.Message, error) {
	ca, err := decodeCatalogArguments(tc.Arguments)
	if err != nil {
		return agent.Message{}, fmt.Errorf("invalid arguments: %w", err)
	}

	cr, err := d.catalogExplorer.Catalog(ca.Group)
	if err != nil {
		return agent.Message{}, fmt.Errorf("failed to explore the catalog: %w", err)
	}

	cJson, err := json.Marshal(cr)
	if err != nil {
		return agent.Message{}, fmt.Errorf("failed to serialize the catalog: %w", err)
	}

	return agent.Message{Role: agent.RoleTool, ToolCallID: tc.ID, Content: string(cJson)}, nil
}
