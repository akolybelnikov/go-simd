package ollama

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/akolybelnikov/go-simd/internal/agent"
)

type modelStream struct {
	body    io.ReadCloser
	decoder *json.Decoder
}

func (s *modelStream) Next() (agent.ModelEvent, error) {
	var resp chatResponse
	if err := s.decoder.Decode(&resp); err != nil {
		return agent.ModelEvent{}, fmt.Errorf("error decoding body: %w", err)
	}
	if resp.Done {
		return agent.ModelEvent{Kind: agent.EventCompleted}, nil
	}

	return agent.ModelEvent{
		Kind: agent.EventTextDelta,
		Text: resp.Message.Content,
	}, nil
}

func (s *modelStream) Close() error {
	return s.body.Close()
}

func newModelStream(body io.ReadCloser) *modelStream {
	return &modelStream{
		body:    body,
		decoder: json.NewDecoder(body),
	}
}
