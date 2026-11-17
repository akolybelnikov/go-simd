package labdispatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DecodeCatalogArguments(t *testing.T) {
	tests := []struct {
		name      string
		arguments []byte
		wantGroup string
		wantErr   bool
	}{
		{
			name:      "rejects unknown field",
			arguments: []byte(`{"group":"addition","command":"anything"}`),
			wantErr:   true,
		},
		{
			name:      "valid",
			arguments: []byte(`{"group":"addition"}`),
			wantGroup: "addition",
		},
		{
			name:      "trailing JSON",
			arguments: []byte(`{"group":"addition"} {"group":"all"}`),
			wantErr:   true,
		},
		{
			name:      "missing group",
			arguments: []byte(`{}`),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := decodeCatalogArguments(tt.arguments)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotZero(t, decoded)
			assert.Equal(t, tt.wantGroup, decoded.Group)
		})
	}
}
