package ollama

import (
	"fmt"
	"net/http"
	"net/url"
)

// Client implements the agent model-client contract using Ollama’s HTTP chat API.
type Client struct {
	baseURL    *url.URL
	modelName  string
	httpClient *http.Client
}

// NewClient creates an Ollama client using the provided base URL, model name, and HTTP client.
func NewClient(baseURL *url.URL, modelName string, httpClient *http.Client) (*Client, error) {
	if baseURL == nil {
		return nil, fmt.Errorf("base URL cannot be nil")
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme %q", baseURL.Scheme)
	}
	if baseURL.Host == "" {
		return nil, fmt.Errorf("base URL host cannot be empty")
	}
	if modelName == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}
	if httpClient == nil {
		return nil, fmt.Errorf("HTTP client cannot be nil")
	}

	urlCp := (*baseURL)

	return &Client{
		baseURL:    &urlCp,
		modelName:  modelName,
		httpClient: httpClient,
	}, nil
}
