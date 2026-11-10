package ollama_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/agent/ollama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	baseURL, err := url.Parse("http://127.0.0.1:11434")
	require.NoError(t, err)

	relativeBaseURL, err := url.Parse("ollama")
	require.NoError(t, err)

	unsupportedSchemeURL, err := url.Parse("ftp://127.0.0.1:11434")
	require.NoError(t, err)

	emptyHostURL, err := url.Parse("http:///ollama")
	require.NoError(t, err)

	tests := []struct {
		name        string
		baseURL     *url.URL
		modelName   string
		httpClient  *http.Client
		wantErrText string
	}{
		{
			name:       "valid arguments",
			baseURL:    baseURL,
			modelName:  "qwen3:4b",
			httpClient: http.DefaultClient,
		},
		{
			name:        "nil base url",
			modelName:   "qwen3:4b",
			httpClient:  http.DefaultClient,
			wantErrText: "base URL cannot be nil",
		},
		{
			name:        "relative base url",
			baseURL:     relativeBaseURL,
			modelName:   "qwen3:4b",
			httpClient:  http.DefaultClient,
			wantErrText: `unsupported URL scheme ""`,
		},
		{
			name:        "unsupported scheme is rejected before other missing arguments",
			baseURL:     unsupportedSchemeURL,
			wantErrText: `unsupported URL scheme "ftp"`,
		},
		{
			name:        "empty host is rejected before other missing arguments",
			baseURL:     emptyHostURL,
			wantErrText: "base URL host cannot be empty",
		},
		{
			name:        "empty model name is rejected before nil http client",
			baseURL:     baseURL,
			wantErrText: "model name cannot be empty",
		},
		{
			name:        "nil http client",
			baseURL:     baseURL,
			modelName:   "qwen3:4b",
			wantErrText: "HTTP client cannot be nil",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ollama.NewClient(tt.baseURL, tt.modelName, tt.httpClient)
			if tt.wantErrText == "" {
				assert.NotZero(t, got)
				assert.NoError(t, gotErr)
			} else {
				assert.Zero(t, got)
				require.ErrorContains(t, gotErr, tt.wantErrText)
			}
		})
	}
}
