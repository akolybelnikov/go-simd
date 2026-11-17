package labdispatch

import (
	"encoding/json"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ agent.ToolDispatcher = (*Dispatcher)(nil)

type fakeCatalogExplorer struct {
	lastRequestedGroup string
	calls              int
}

func (c *fakeCatalogExplorer) Catalog(group string) (labapi.CatalogResponse, error) {
	c.calls++
	c.lastRequestedGroup = group

	return labapi.CatalogResponse{Schema: labapi.CatalogSchema}, nil
}

func TestDispatcher_ExploreCatalog(t *testing.T) {
	tests := []struct {
		name      string
		group     string
		arguments []byte
		wantErr   bool
	}{
		{name: "addition", group: "addition", arguments: []byte(`{"group":"addition"}`)},
		{name: "multiplication", group: "multiplication", arguments: []byte(`{"group":"multiplication"}`)},
		{name: "rejects invalid catalog arguments", arguments: []byte(`{}`), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &fakeCatalogExplorer{}
			d, err := NewDispatcher(c)
			require.NoError(t, err)

			tc := agent.ToolCall{ID: "call-1", Name: "explore_catalog", Arguments: tt.arguments}
			msg, err := d.Dispatch(t.Context(), tc)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, agent.Message{}, msg)
				assert.Zero(t, c.calls)
				assert.Empty(t, c.lastRequestedGroup)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, 1, c.calls)
			assert.Equal(t, tt.group, c.lastRequestedGroup)
			assert.Equal(t, agent.RoleTool, msg.Role)
			assert.Equal(t, "call-1", msg.ToolCallID)

			var cr labapi.CatalogResponse
			err = json.Unmarshal([]byte(msg.Content), &cr)
			require.NoError(t, err)
			assert.Equal(t, labapi.CatalogSchema, cr.Schema)
			require.NoError(t, msg.Validate())
		})
	}
}
