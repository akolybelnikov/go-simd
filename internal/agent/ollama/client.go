package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/agent"
)

// Client implements the agent model-client contract using Ollama’s HTTP chat API.
type Client struct {
	baseURL    *url.URL
	modelName  string
	httpClient *http.Client
}

// Stream sends a chat request to Ollama and opens a stream of model events.
func (c *Client) Stream(ctx context.Context, req agent.ModelRequest) (agent.ModelStream, error) {
	chatReq, err := c.buildChatRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build a chat request: %w", err)
	}

	encoded, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode a chat request: %w", err)
	}

	ep := c.baseURL.JoinPath("api", "chat")
	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		ep.String(),
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create an http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		r := io.LimitReader(resp.Body, 4<<10)
		respBytes, readErr := io.ReadAll(r)
		closeErr := resp.Body.Close()

		if readErr != nil {
			return nil, fmt.Errorf("error reading the response body %w with http status: %d", readErr, resp.StatusCode)
		}

		if closeErr != nil {
			return nil, fmt.Errorf("error closing the response body %w with http status: %d", closeErr, resp.StatusCode)
		}

		respStr := strings.TrimSpace(string(respBytes))
		return nil, fmt.Errorf("unexpected response %s with http status: %d", respStr, resp.StatusCode)
	}

	return newModelStream(resp.Body), nil
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
